// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_connection

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	"data-model/logics"
	"data-model/logics/data_connection/data_source"
)

var (
	dcsOnce sync.Once
	dcs     interfaces.DataConnectionService
)

type dataConnectionService struct {
	appSetting *common.AppSetting
	db         *sql.DB
	dca        interfaces.DataConnectionAccess
}

func NewDataConnectionService(appSetting *common.AppSetting) interfaces.DataConnectionService {
	dcsOnce.Do(func() {
		dcs = &dataConnectionService{
			appSetting: appSetting,
			db:         logics.DB,
			dca:        logics.DCA,
		}
	})
	return dcs
}

func (dcs *dataConnectionService) CreateDataConnection(ctx context.Context, conn *interfaces.DataConnection) (connID string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 创建数据连接")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	processor, err := data_source.NewDataConnectionProcessor(ctx, dcs.appSetting, conn.DataSourceType)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return "", err
	}

	// 1. 创建前的校验
	err = processor.ValidateWhenCreate(ctx, conn)
	if err != nil {
		return "", err
	}

	// 2. 获取详细配置的md5
	md5, err := processor.ComputeConfigMD5(ctx, conn)
	if err != nil {
		return "", err
	}
	// 更新md5
	conn.DataSourceConfigMD5 = md5

	// 3. 根据md5, 检查数据库中是否有重复的详细配置
	connMap, err := dcs.getDataConnectionsByConfigMD5(ctx, md5)
	if err != nil {
		return "", err
	}

	if len(connMap) > 0 {
		for id := range connMap {
			errDetails := fmt.Sprintf("Same config whose conn_id in %s already exists in the database", id)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest,
				derrors.DataModel_DataConnection_DuplicatedParameter_Config).WithErrorDetails(errDetails)
		}
	}

	// 4. 获取完整的auth_info和连接状态
	err = processor.GenerateAuthInfoAndStatus(ctx, conn)
	if err != nil {
		return "", err
	}

	// 5. 生成分布式ID
	conn.ID = xid.New().String()

	conn.DataConnectionStatus.ID = conn.ID

	// 6. 生成update_time
	conn.CreateTime = time.Now().UnixMilli()
	conn.UpdateTime = conn.CreateTime

	// 7. 开始事务
	tx, err := dcs.db.Begin()
	if err != nil {
		errDetails := fmt.Sprintf("Begin transaction failed when creating a data connection, err: %v", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_InternalError_BeginTransactionFailed).
			WithErrorDetails(errDetails)
	}

	defer func() {
		if err != nil {
			// 回滚事务
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				errDetails := fmt.Sprintf("Rollback transaction failed when creating a data connection, err: %v", rollbackErr)
				logger.Errorf(errDetails)
				o11y.Error(ctx, errDetails)
			}
		} else {
			// 提交事务
			commitErr := tx.Commit()
			if commitErr != nil {
				errDetails := fmt.Sprintf("Commit transaction failed when creating a data connection, err: %v", commitErr)
				logger.Errorf(errDetails)
				o11y.Error(ctx, errDetails)
				err = rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_InternalError_CommitTransactionFailed).WithErrorDetails(errDetails)
			}
		}
	}()

	// 8. 创建数据连接
	err = dcs.dca.CreateDataConnection(ctx, tx, conn)
	if err != nil {
		logger.Errorf("Create a data connection failed, err: %v", err.Error())
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_CreateDataConnectionFailed).WithErrorDetails(err.Error())
	}

	// 9. 创建数据连接状态
	err = dcs.dca.CreateDataConnectionStatus(ctx, tx, conn.DataConnectionStatus)
	if err != nil {
		logger.Errorf("Create a data connection status failed, err: %v", err.Error())
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_CreateDataConnectionStatusFailed).WithErrorDetails(err.Error())
	}

	return conn.ID, nil
}

func (dcs *dataConnectionService) DeleteDataConnections(ctx context.Context, connIDs []string) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 批量删除数据连接")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 开始事务
	tx, err := dcs.db.Begin()
	if err != nil {
		errDetails := fmt.Sprintf("Begin transaction failed when deleting data connections, err: %v", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_InternalError_BeginTransactionFailed).WithErrorDetails(err.Error())
	}

	defer func() {
		if err != nil {
			// 回滚事务
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				errDetails := fmt.Sprintf("Rollback transaction failed when deleting data connections, err: %v", rollbackErr)
				logger.Error(errDetails)
				o11y.Error(ctx, errDetails)
			}
		} else {
			// 提交事务
			commitErr := tx.Commit()
			if commitErr != nil {
				errDetails := fmt.Sprintf("Commit transaction failed when deleting data connections, err: %v", commitErr)
				logger.Errorf(errDetails)
				o11y.Error(ctx, errDetails)
				err = rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_InternalError_CommitTransactionFailed).WithErrorDetails(errDetails)
			}
		}
	}()

	// 2. 删除数据连接
	err = dcs.dca.DeleteDataConnections(ctx, tx, connIDs)
	if err != nil {
		logger.Errorf("Delete data connections failed, err: %v", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_DeleteDataConnectionsFailed).WithErrorDetails(err.Error())
	}

	// 3. 删除数据连接状态
	err = dcs.dca.DeleteDataConnectionStatuses(ctx, tx, connIDs)
	if err != nil {
		logger.Errorf("Delete data connection statuses failed, err: %v", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_DeleteDataConnectionStatusesFailed).WithErrorDetails(err.Error())
	}

	return nil
}

func (dcs *dataConnectionService) UpdateDataConnection(ctx context.Context,
	conn *interfaces.DataConnection, preConn *interfaces.DataConnection) (err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 修改数据连接")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	processor, err := data_source.NewDataConnectionProcessor(ctx, dcs.appSetting, conn.DataSourceType)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	// 1. 修改前的校验
	err = processor.ValidateWhenUpdate(ctx, conn, preConn)
	if err != nil {
		return err
	}

	// 2. 获取详细配置的md5
	md5, err := processor.ComputeConfigMD5(ctx, conn)
	if err != nil {
		return err
	}
	conn.DataSourceConfigMD5 = md5

	// 3. 判断配置是否变更, 若变更, 则另需其它操作
	if md5 != preConn.DataSourceConfigMD5 {
		// 3.1 根据md5, 检查数据库中是否有重复的详细配置
		connMap, err := dcs.getDataConnectionsByConfigMD5(ctx, md5)
		if err != nil {
			return err
		}

		if len(connMap) > 0 {
			for id := range connMap {
				errDetails := fmt.Sprintf("Same config whose conn_id is %s already exists in the database", id)
				logger.Error(errDetails)
				o11y.Error(ctx, errDetails)
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					derrors.DataModel_DataConnection_DuplicatedParameter_Config).WithErrorDetails(errDetails)
			}
		}
	}

	// 4. 获取完整的auth_info和连接状态
	err = processor.GenerateAuthInfoAndStatus(ctx, conn)
	if err != nil {
		return err
	}

	// 5. 生成update_time
	conn.UpdateTime = time.Now().UnixMilli()

	// 6. 更新配置
	err = dcs.updateDataConnectionAndStatus(ctx, conn)
	if err != nil {
		return err
	}
	return nil
}

func (dcs *dataConnectionService) GetDataConnection(ctx context.Context, conID string,
	withAuthInfo bool) (conn *interfaces.DataConnection, isExist bool, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 查询数据连接")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 获取数据连接详情
	conn, isExist, err = dcs.dca.GetDataConnection(ctx, conID)
	if err != nil {
		logger.Errorf("Get data connection failed, err: %v", err.Error())
		return conn, isExist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_GetDataConnectionsFailed).WithErrorDetails(err.Error())
	}

	if !isExist {
		return conn, isExist, nil
	}

	// 2. 对详情进行处理
	processor, err := data_source.NewDataConnectionProcessor(ctx, dcs.appSetting, conn.DataSourceType)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return conn, isExist, err
	}

	// 3. 更新auth_info和connection_status
	// 这里的err暂时不处理, 防止前面页面编辑前一直无法获取DataConnection详情
	needWriteBack, _ := processor.UpdateAuthInfoAndStatus(ctx, conn)

	// 4. 判断是否需要回写数据库
	if needWriteBack {
		err = dcs.updateDataConnectionAndStatus(ctx, conn)
		if err != nil {
			return conn, isExist, err
		}
	}

	// 5. 判断是否需要隐藏auth_info
	if !withAuthInfo {
		err = processor.HideAuthInfo(ctx, conn)
		if err != nil {
			return conn, isExist, err
		}
	}

	return conn, isExist, nil
}

func (dcs *dataConnectionService) ListDataConnections(ctx context.Context,
	queryParams interfaces.DataConnectionListQueryParams) (entries []*interfaces.DataConnectionListEntry, total int, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 查询数据连接列表与总数")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 获取数据连接列表
	entries, err = dcs.dca.ListDataConnections(ctx, queryParams)
	if err != nil {
		logger.Errorf("List data connections failed, err: %v", err.Error())
		return entries, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_ListDataConnectionsFailed).WithErrorDetails(err.Error())
	}

	// 2. 获取数据连接总数
	total, err = dcs.dca.GetDataConnectionTotal(ctx, queryParams)
	if err != nil {
		logger.Errorf("Get data connection total failed, err: %v", err.Error())
		return entries, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_GetDataConnectionTotalFailed).WithErrorDetails(err.Error())
	}

	return entries, total, nil
}

func (dcs *dataConnectionService) GetMapAboutName2ID(ctx context.Context, connNames []string) (name2ID map[string]string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 查询数据连接名称与ID的映射关系")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	name2ID, err = dcs.dca.GetMapAboutName2ID(ctx, connNames)
	if err != nil {
		logger.Errorf("Get data connection map about name to id failed, err: %v", err.Error())
		return name2ID, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_GetMapAboutName2IDFailed).WithErrorDetails(err.Error())
	}
	return name2ID, nil
}

func (dcs *dataConnectionService) GetMapAboutID2Name(ctx context.Context, connIDs []string) (id2Name map[string]string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 查询数据连接ID与名称的映射关系")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	id2Name, err = dcs.dca.GetMapAboutID2Name(ctx, connIDs)
	if err != nil {
		logger.Errorf("Get data connection map about id to name failed, err: %v", err.Error())
		return id2Name, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_GetMapAboutID2NameFailed).WithErrorDetails(err.Error())
	}
	return id2Name, nil
}

func (dcs *dataConnectionService) GetDataConnectionSourceType(ctx context.Context, connID string) (sourceType string, isExist bool, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 根据数据连接ID查询数据来源类型")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	sourceType, isExist, err = dcs.dca.GetDataConnectionSourceType(ctx, connID)
	if err != nil {
		logger.Errorf("Get data connection data_source_type failed, err: %v", err.Error())
		return sourceType, isExist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_GetDataConnectionSourceTypeFailed).WithErrorDetails(err.Error())
	}

	return sourceType, isExist, nil
}

/*
	私有方法
*/

func (dcs *dataConnectionService) updateDataConnectionAndStatus(ctx context.Context, conn *interfaces.DataConnection) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 更新数据库中数据连接配置和状态")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 开始事务
	tx, err := dcs.db.Begin()
	if err != nil {
		errDetails := fmt.Sprintf("Begin transaction failed when updating data connection config and status in database, err: %v", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_InternalError_BeginTransactionFailed).
			WithErrorDetails(errDetails)
	}

	defer func() {
		if err != nil {
			// 回滚事务
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				errDetails := fmt.Sprintf("Rollback transaction failed when updating data connection config and status in database, err: %v", rollbackErr)
				logger.Errorf(errDetails)
				o11y.Error(ctx, errDetails)
			}
		} else {
			// 提交事务
			commitErr := tx.Commit()
			if commitErr != nil {
				errDetails := fmt.Sprintf("Commit transaction failed when updating data connection config and status in database, err: %v", commitErr)
				logger.Errorf(errDetails)
				o11y.Error(ctx, errDetails)
				err = rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_InternalError_CommitTransactionFailed).WithErrorDetails(errDetails)
			}
		}
	}()

	// 3. 修改数据连接
	err = dcs.dca.UpdateDataConnection(ctx, tx, conn)
	if err != nil {
		logger.Errorf("Update a data connection failed, err: %v", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_UpdateDataConnectionFailed).WithErrorDetails(err.Error())
	}

	// 4. 修改数据连接状态
	err = dcs.dca.UpdateDataConnectionStatus(ctx, tx, conn.DataConnectionStatus)
	if err != nil {
		logger.Errorf("Update a data connection status failed, err: %v", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_UpdateDataConnectionStatusFailed).WithErrorDetails(err.Error())
	}

	return nil
}

func (dcs *dataConnectionService) getDataConnectionsByConfigMD5(ctx context.Context,
	configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 根据详细配置的md5从查询数据连接")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	connMap, err = dcs.dca.GetDataConnectionsByConfigMD5(ctx, configMD5)
	if err != nil {
		logger.Errorf("Get data connections by data_source_type failed, err: %v", err.Error())
		return connMap, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_GetDataConnectionsFailed).WithErrorDetails(err.Error())
	}

	return connMap, nil
}
