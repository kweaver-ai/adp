// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

// import (
// 	"context"
// 	"crypto/rand"
// 	"database/sql"
// 	"errors"
// 	"fmt"
// 	"net/http"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/kweaver-ai/kweaver-go-lib/logger"
// 	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
// 	"github.com/kweaver-ai/kweaver-go-lib/rest"
// 	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
// 	"github.com/jinzhu/copier"
// 	"github.com/rs/xid"
// 	attr "go.opentelemetry.io/otel/attribute"
// 	"go.opentelemetry.io/otel/codes"
// 	"go.uber.org/zap"

// 	"data-model/common"
// 	derrors "data-model/errors"
// 	"data-model/interfaces"
// 	"data-model/logics"
// 	"data-model/logics/permission"
// )

// var (
// 	dsServiceOnce sync.Once
// 	dsService     interfaces.DataSourceService
// )

// type dataSourceService struct {
// 	appSetting *common.AppSetting
// 	db         *sql.DB
// 	dsa        interfaces.DataSourceAccess
// 	iba        interfaces.IndexBaseAccess
// 	sra        interfaces.ScanRecordAccess
// 	vga        interfaces.VegaGatewayAccess
// 	vma        interfaces.VegaMetadataAccess
// 	dvs        interfaces.DataViewService
// 	ps         interfaces.PermissionService
// }

// func NewDataSourceService(appSetting *common.AppSetting) interfaces.DataSourceService {
// 	dsServiceOnce.Do(func() {
// 		dsService = &dataSourceService{
// 			appSetting: appSetting,
// 			db:         logics.DB,
// 			dsa:        logics.DSA,
// 			iba:        logics.IBA,
// 			sra:        logics.SRA,
// 			vga:        logics.VGA,
// 			vma:        logics.VMA,
// 			dvs:        NewDataViewService(appSetting),
// 			ps:         permission.NewPermissionService(appSetting),
// 		}
// 	})

// 	return dsService
// }

// // 扫描单个数据源
// func (dss *dataSourceService) Scan(ctx context.Context, req *interfaces.ScanTask) (*interfaces.ScanResult, error) {
// 	// if mode == interfaces.ScanMode_Concurrent {
// 	// 	return dss.ConcurrentScan(ctx, req)
// 	// }
// 	return dss.SerialScan(ctx, req)
// }

// // 获取带有扫描记录的数据源列表
// func (dss *dataSourceService) ListDataSourcesWithScanRecord(ctx context.Context,
// 	queryParams *interfaces.ListDataSourceQueryParams) (*interfaces.ListDataSourcesResult, error) {
// 	// 请求数据源的接口获取数据源列表
// 	dataSources, err := dss.dsa.ListDataSources(ctx)
// 	if err != nil {
// 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 			WithErrorDetails(err.Error())
// 	}

// 	// 获取所有的扫描记录
// 	scanRecords, err := dss.sra.ListScanRecords(ctx, &interfaces.PaginationQueryParameters{Limit: -1})
// 	if err != nil {
// 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_InternalError_DataBaseError).
// 			WithErrorDetails(err.Error())
// 	}

// 	// 遍历数据源列表，将扫描记录添加到数据源中，按照数据源类型分成两类
// 	DSLDataSources := []*interfaces.DataSource{}
// 	SQLDataSources := []*interfaces.DataSource{}
// 	for _, ds := range dataSources.Entries {
// 		if ds.Type == interfaces.DataSourceType_Excel || ds.Type == interfaces.DataSourceType_IndexBase {
// 			ds.LastScanTime = time.Now().UnixMilli()
// 			ds.Status = interfaces.DataSourceAvailable
// 		}

// 		for _, sr := range scanRecords {
// 			if ds.ID == sr.DataSourceID {
// 				ds.Status = sr.DataSourceStatus
// 				ds.LastScanTime = sr.ScanTime
// 				break
// 			}
// 		}

// 		if ds.Type == interfaces.DataSourceType_IndexBase {
// 			DSLDataSources = append(DSLDataSources, ds)
// 		} else {
// 			SQLDataSources = append(SQLDataSources, ds)
// 		}
// 	}

// 	if strings.ToUpper(queryParams.QueryType) == interfaces.QueryType_DSL {
// 		dataSources.Entries = DSLDataSources
// 		dataSources.TotalCount = len(DSLDataSources)
// 	} else if strings.ToUpper(queryParams.QueryType) == interfaces.QueryType_SQL {
// 		dataSources.Entries = SQLDataSources
// 		dataSources.TotalCount = len(SQLDataSources)
// 	}

// 	return dataSources, nil
// }

// func (dss *dataSourceService) SerialScan(ctx context.Context, req *interfaces.ScanTask) (res *interfaces.ScanResult, err error) {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "SerialScan")
// 	defer span.End()

// 	// 检查用户的权限
// 	// 判断userid是否有创建和更新视图的权限（策略决策）
// 	err = dss.ps.CheckPermission(ctx,
// 		interfaces.Resource{
// 			Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
// 			ID:   interfaces.RESOURCE_ID_ALL,
// 		},
// 		[]string{interfaces.OPERATION_TYPE_CREATE},
// 	)

// 	if err != nil {
// 		span.SetStatus(codes.Error, "SerialScan CheckPermission failed")
// 		return nil, err
// 	}

// 	res, err = dss.SerialScanDataSource(ctx, req)
// 	if err != nil {
// 		span.SetStatus(codes.Error, "SerialScan SerialScanDataSource failed")
// 		return nil, err
// 	}

// 	span.SetStatus(codes.Ok, "SerialScan success")
// 	return res, nil
// }

// func (dss *dataSourceService) SerialScanDataSource(ctx context.Context, req *interfaces.ScanTask) (*interfaces.ScanResult, error) {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Serial scan data source")
// 	defer span.End()

// 	// 新建一个ctx, 使超时不受前端影响
// 	ctx = context.WithoutCancel(ctx)

// 	res := &interfaces.ScanResult{}
// 	//获取数据源信息
// 	dataSource, err := dss.dsa.GetDataSourceByID(ctx, req.DataSourceID)
// 	if err != nil {
// 		span.SetStatus(codes.Error, "get data source by id failed")
// 		o11y.Error(ctx, fmt.Sprintf("get data source by id failed: %s", err.Error()))
// 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 			WithErrorDetails("get data source by id failed")
// 	}

// 	if dataSource.Type == interfaces.DataSourceType_Excel || dataSource.Type == interfaces.DataSourceType_IndexBase ||
// 		dataSource.Type == interfaces.DataSourceType_TingYun || dataSource.Type == interfaces.DataSourceType_AS7 {
// 		span.SetStatus(codes.Error, "excel-type data source do not support scan")
// 		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
// 			WithErrorDetails(fmt.Sprintf("%s type data sources do not support scanning", dataSource.Type))
// 	}

// 	// data_source,status从scan_record表获取
// 	record, exist, err := dss.sra.GetByDataSourceId(ctx, dataSource.ID)
// 	if err != nil {
// 		span.SetStatus(codes.Error, "get scan record failed")
// 		o11y.Error(ctx, fmt.Sprintf("get scan record failed: %s", err.Error()))
// 		logger.Errorf("get scan record Error, %v", err)
// 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 			WithErrorDetails(fmt.Sprintf("get scan record failed, %v", err))
// 	}

// 	if exist && record.DataSourceStatus == interfaces.DataSourceScanning {
// 		span.SetStatus(codes.Error, "data source is scanning already")
// 		o11y.Error(ctx, "data source is scanning already")
// 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 			WithErrorDetails("data source is scanning already")
// 	}

// 	if !exist {
// 		record = &interfaces.ScanRecord{
// 			RecordID:         xid.New().String(),
// 			DataSourceID:     dataSource.ID,
// 			Scanner:          interfaces.ManagementScanner,
// 			ScanTime:         time.Now().UnixMilli(),
// 			DataSourceStatus: interfaces.DataSourceScanning,
// 		}
// 		err = dss.sra.CreateScanRecord(ctx, record)
// 		if err != nil {
// 			span.SetStatus(codes.Error, "create scan record failed")
// 			o11y.Error(ctx, fmt.Sprintf("create scan record failed: %s", err.Error()))
// 			logger.Errorf("updateScanRecord Create Error", zap.Error(err))
// 			return res, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 				WithErrorDetails("create scan record failed")
// 		}
// 	}

// 	//设置数据源扫描中
// 	record.DataSourceStatus = interfaces.DataSourceScanning
// 	if err = dss.sra.UpdateScanRecordStatus(ctx, &interfaces.ScanRecordStatus{
// 		ID:     record.RecordID,
// 		Status: interfaces.DataSourceScanning,
// 	}); err != nil {
// 		span.SetStatus(codes.Error, "update scan record status failed")
// 		o11y.Error(ctx, fmt.Sprintf("update scan record status failed: %s", err.Error()))
// 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 			WithErrorDetails(fmt.Sprintf("update data source status failed, %v", err))
// 	}
// 	defer func() {
// 		//设置数据源可用
// 		record.DataSourceStatus = interfaces.DataSourceAvailable
// 		if err = dss.sra.UpdateScanRecordStatus(ctx, &interfaces.ScanRecordStatus{
// 			ID:     record.RecordID,
// 			Status: interfaces.DataSourceAvailable,
// 		}); err != nil {
// 			span.SetStatus(codes.Error, "update scan record status failed")
// 			o11y.Error(ctx, fmt.Sprintf("update scan record status failed: %s", err.Error()))
// 			logger.Errorf("[ScanDataSource] UpdateDataSource DataSourceAvailable Error, %v", err)
// 		}
// 	}()

// 	//采集元数据
// 	if err = dss.Collect(ctx, dataSource); err != nil {
// 		span.SetStatus(codes.Error, "collect data source failed")
// 		o11y.Error(ctx, fmt.Sprintf("collect data source failed: %s", err.Error()))
// 		logger.Errorf("ScanDataSource] Scan Collect Error, %v", err)
// 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 			WithErrorDetails(fmt.Sprintf("collect data source failed, %v", err))
// 	}

// 	data, err := dss.GetDataSourceAllTableInfo(ctx, dataSource)
// 	if err != nil {
// 		span.SetStatus(codes.Error, "get data source table info failed")
// 		o11y.Error(ctx, fmt.Sprintf("get data source table info failed: %s", err.Error()))
// 		logger.Errorf("[ScanDataSource] UpdateDataSource DataSourceAvailable Error, %v", err)
// 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 			WithErrorDetails("get data source table info failed")
// 	}

// 	o11y.Info(ctx, fmt.Sprintf("get table count %d ", len(data)))
// 	logger.Infof("[vma] get table count %d ", len(data))

// 	// dataViewSource := dataSource.BinData.DataViewSource
// 	// if len(data) == 0 && dataViewSource == "" {
// 	// excel 类 metadata 是否有？需要测试
// 	if len(data) == 0 {
// 		span.SetStatus(codes.Error, "data source has no table")
// 		o11y.Error(ctx, "data source has no table")
// 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 			WithErrorDetails("data source has no table")
// 	}

// 	err = dss.SerialCompareView(ctx, dataSource, data, res)
// 	if err != nil {
// 		span.SetStatus(codes.Error, "serial compare view failed")
// 		o11y.Error(ctx, fmt.Sprintf("serial compare view failed: %s", err.Error()))
// 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 			WithErrorDetails(fmt.Sprintf("serial compare view failed, %v", err))
// 	}

// 	//扫描成功，更新扫描记录
// 	if err = dss.updateScanRecord(ctx, dataSource.ID, req.TaskID, req.ProjectID); err != nil {
// 		span.SetStatus(codes.Error, "update scan record failed")
// 		o11y.Error(ctx, fmt.Sprintf("update scan record failed: %s", err.Error()))
// 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 			WithErrorDetails(fmt.Sprintf("update scan record failed, %v", err))
// 	}

// 	span.SetStatus(codes.Ok, "scan data source success")
// 	return res, nil
// }

// type VETimeCost struct {
// 	createViewCost  time.Duration
// 	createViewMax   time.Duration
// 	createViewCount int
// 	updateViewCost  time.Duration
// 	updateViewMax   time.Duration
// 	updateViewCount int
// }

// func (dss *dataSourceService) CostCalculate(costCh chan *VETimeCost, cost *VETimeCost) {
// 	for onePageCost := range costCh {
// 		cost.createViewCost += onePageCost.createViewCost
// 		cost.createViewCount += onePageCost.createViewCount
// 		cost.updateViewCost += onePageCost.updateViewCost
// 		cost.updateViewCount += onePageCost.updateViewCount
// 		if cost.createViewMax < onePageCost.createViewMax {
// 			cost.createViewMax = onePageCost.createViewMax
// 		}
// 		if cost.updateViewMax < onePageCost.updateViewMax {
// 			cost.updateViewMax = onePageCost.updateViewMax
// 		}
// 	}
// }

// func (dss *dataSourceService) ReceiveErrorView(ctx context.Context, ch chan *interfaces.ErrorView, createViewError *[]*interfaces.ErrorView) {
// 	for re := range ch {
// 		*createViewError = append(*createViewError, re)
// 	}
// }

// func (dss *dataSourceService) Receive(ctx context.Context, createViewErrorCh chan *interfaces.ErrorView, createViewError *[]*interfaces.ErrorView) {
// 	for re := range createViewErrorCh {
// 		*createViewError = append(*createViewError, re)
// 	}
// }

// func (dss *dataSourceService) PollingCollectTask2(ctx context.Context, dataSource *interfaces.DataSource) error {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Polling collect task")
// 	defer span.End()

// 	record, _, err := dss.sra.GetByDataSourceId(ctx, dataSource.ID)
// 	if err != nil {
// 		span.SetStatus(codes.Error, "get scan record failed")
// 		logger.Errorf("get scan record Error, %v", err)
// 		return err
// 	}

// 	taskId := record.MetadataTaskID
// 	for i := 1; true; i++ {
// 		tasks, taskErr := dss.vma.GetTasks(ctx, &interfaces.GetTasksReq{Keyword: taskId})
// 		if taskErr != nil {
// 			span.SetStatus(codes.Error, "get tasks failed")
// 			logger.Errorf("[Scan Collect] GetTasks error :%s\n", taskErr)
// 			return taskErr
// 		}

// 		if len(tasks.Data) == 1 && tasks.Data[0].Status == 2 {
// 			span.SetStatus(codes.Ok, "collect task retry")
// 			logger.Infof("[Scan Collect]  retry dataSource: %s,times: #%d\n", dataSource.ID, i)
// 			time.Sleep(time.Second * time.Duration(i*10))
// 			continue
// 		}

// 		if len(tasks.Data) == 1 && tasks.Data[0].Status == 1 {
// 			span.SetStatus(codes.Error, "collect task failed")
// 			logger.Infof("[Scan Collect] TaskFail, taskId :%s, status :%d", taskId, tasks.Data[0].Status)
// 			return errors.New("metadata CollectTask failed")
// 		}

// 		span.SetStatus(codes.Ok, "collect task finished")
// 		logger.Infof("[Scan Collect] TaskFinish, taskId :%s, status :%d", taskId, tasks.Data[0].Status)
// 		return nil
// 	}

// 	span.SetStatus(codes.Ok, "")
// 	return nil
// }

// func (dss *dataSourceService) GetDataSourceAllTableInfo(ctx context.Context, dataSource *interfaces.DataSource) (
// 	[]*interfaces.GetDataTableDetailDataBatchRes, error) {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get data source all table info")
// 	defer span.End()

// 	tableDetail, err := dss.vma.GetDataTableDetailBatch(ctx, &interfaces.GetDataTableDetailBatchReq{
// 		Limit:        1000,
// 		Offset:       1,
// 		DataSourceId: dataSource.ID,
// 	})
// 	if err != nil {
// 		span.SetStatus(codes.Error, "get data table detail failed")
// 		o11y.Error(ctx, "get data table detail failed")
// 		return nil, err
// 	}

// 	if tableDetail.TotalCount <= 1000 {
// 		span.SetStatus(codes.Ok, "get data table detail finished, total count < 1000")
// 		return tableDetail.Data, nil
// 	}

// 	res := make([]*interfaces.GetDataTableDetailDataBatchRes, 0)
// 	res = append(res, tableDetail.Data...)
// 	for i := 0; i < tableDetail.TotalCount/1000; i++ {
// 		nextTable, err := dss.vma.GetDataTableDetailBatch(ctx, &interfaces.GetDataTableDetailBatchReq{
// 			Limit:        1000,
// 			Offset:       i + 2,
// 			DataSourceId: dataSource.ID,
// 		})
// 		if err != nil {
// 			span.SetStatus(codes.Error, "get data table detail failed")
// 			o11y.Error(ctx, "get data table detail failed")
// 			return nil, err
// 		}
// 		res = append(res, nextTable.Data...)
// 	}

// 	span.SetStatus(codes.Ok, "get data table detail finished, total count > 1000")
// 	return res, nil
// }

// func (dss *dataSourceService) Collect(ctx context.Context, datasource *interfaces.DataSource) error {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Collect metadata")
// 	defer span.End()

// 	collectRes, err := dss.vma.DoCollect(ctx, &interfaces.DoCollectReq{DataSourceId: datasource.ID})
// 	if err != nil {
// 		span.SetStatus(codes.Error, "collect metadata failed")
// 		logger.Errorf("[ScanDataSource] Scan Collect DoCollect Error, %v", err)
// 		return err
// 	}
// 	var taskId string
// 	split := strings.Split(collectRes.Data, "任务ID:")
// 	if len(split) == 2 {
// 		split2 := strings.Split(split[1], "}")
// 		if len(split2) > 0 {
// 			taskId = split2[0]
// 		}
// 	}
// 	if taskId == "" || len(taskId) != 19 {
// 		span.SetStatus(codes.Error, "get task id failed")
// 		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 			WithErrorDetails(fmt.Sprintf("get task id failed, data = %v", collectRes.Data))
// 	}
// 	o11y.Info(ctx, "collect metadata success")
// 	logger.Infof("DoCollect taskId :%s", taskId)

// 	// 适配dip data_source,status从scan_record表获取
// 	record, _, err := dss.sra.GetByDataSourceId(ctx, datasource.ID)
// 	if err != nil {
// 		span.SetStatus(codes.Error, "get scan record failed")
// 		logger.Errorf("get scan record Error, %v", err)
// 		return err
// 	}
// 	record.MetadataTaskID = taskId
// 	if err = dss.sra.UpdateScanRecord(ctx, record); err != nil {
// 		span.SetStatus(codes.Error, "update scan record failed")
// 		o11y.Error(ctx, "update scan record failed")
// 		logger.Errorf("[ScanDataSource] databaseError update metadataTaskId failed, %v", err)
// 	}

// 	defer func() {
// 		record.MetadataTaskID = taskId
// 		if err = dss.sra.UpdateScanRecord(ctx, record); err != nil {
// 			span.SetStatus(codes.Error, "update scan record failed")
// 			o11y.Error(ctx, "update scan record failed")
// 			logger.Errorf("[ScanDataSource] DatabaseError update meta data task id failed, %v", err)
// 		}
// 	}()

// 	if err = dss.PollingCollectTask2(ctx, datasource); err != nil {
// 		span.SetStatus(codes.Error, "polling collect task failed")
// 		o11y.Error(ctx, "polling collect task failed")
// 		return err
// 	}

// 	span.SetStatus(codes.Ok, "collect metadata finished")
// 	return nil
// }

// func (dss *dataSourceService) SerialCompareView(ctx context.Context, dataSource *interfaces.DataSource,
// 	tables []*interfaces.GetDataTableDetailDataBatchRes, res *interfaces.ScanResult) error {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Serial Compare View")
// 	defer span.End()

// 	// 获取这个数据源（分组）下的所有视图
// 	dataViews, err := dss.dvs.GetDataViewsBySourceID(ctx, dataSource.ID)
// 	if err != nil {
// 		span.SetStatus(codes.Error, "get data views by source id failed")
// 		o11y.Error(ctx, "get data views by source id failed")
// 		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 			WithErrorDetails(err.Error())
// 	}
// 	dataViewsMap := make(map[string]*FormViewFlag)
// 	// 维护业务名称map，避免新生成的业务名称会在分组内重复
// 	dataViewBusinessNameMap := make(map[string]struct{})
// 	for _, dView := range dataViews {
// 		dataViewsMap[dView.TechnicalName] = &FormViewFlag{DataView: dView, flag: 1}
// 		dataViewBusinessNameMap[dView.ViewName] = struct{}{}
// 	}

// 	tablesMap := make(map[string]*interfaces.GetDataTableDetailDataBatchRes)
// 	for _, table := range tables {
// 		tablesMap[table.Name] = table
// 	}

// 	// delete view
// 	deleteViewIDs := make([]string, 0)
// 	for techName, vv := range dataViewsMap {
// 		if _, ok := tablesMap[techName]; !ok {
// 			// 源表被删除了，标记为源表删除
// 			deleteViewIDs = append(deleteViewIDs, vv.ViewID)

// 		}
// 	}

// 	// 更新源表删除的视图状态
// 	err = dss.MarkViewAsSourceDeleted(ctx, deleteViewIDs)
// 	if err != nil {
// 		span.SetStatus(codes.Error, "mark view as source deleted failed")
// 		o11y.Error(ctx, "mark view as source deleted failed")
// 		return err
// 	}

// 	//统计失败视图
// 	createViewErrorCh := make(chan *interfaces.ErrorView)
// 	createViewError := make([]*interfaces.ErrorView, 0)
// 	go dss.ReceiveErrorView(ctx, createViewErrorCh, &createViewError)

// 	createViewCtx, createSpan := ar_trace.Tracer.Start(ctx, "Create data views")
// 	createSpan.SetAttributes(
// 		attr.Key("data_source_id").String(dataSource.ID),
// 		attr.Key("data_source_name").String(dataSource.Name))
// 	defer createSpan.End()

// 	// 对比找出需要创建和更新的
// 	needCreateViews := make([]*interfaces.GetDataTableDetailDataBatchRes, 0, len(tables))
// 	for _, table := range tables {
// 		tablesMap[table.Name] = table

// 		// 表名对应视图的技术名称
// 		if existTable, ok := dataViewsMap[table.Name]; !ok {
// 			needCreateViews = append(needCreateViews, table)

// 		} else {
// 			// 更新视图接口不支持批量
// 			if err = dss.updateView(createViewCtx, dataSource, existTable.DataView, table); err != nil {
// 				createSpan.SetStatus(codes.Error, "update view failed")
// 				o11y.Error(createViewCtx, "update view failed")
// 				return err
// 			}
// 			dataViewsMap[table.Name].flag = 2
// 		}
// 	}

// 	// 批量创建视图
// 	if err = dss.createViews(createViewCtx, dataSource, needCreateViews, dataViewBusinessNameMap); err != nil {
// 		createSpan.SetStatus(codes.Error, "create view failed")
// 		o11y.Error(createViewCtx, "create view failed")
// 		return err
// 	}

// 	createSpan.SetStatus(codes.Ok, "create views success")

// 	close(createViewErrorCh)
// 	res.ErrorView = createViewError
// 	res.ErrorViewCount = len(createViewError)
// 	res.ScanViewCount = len(tables)

// 	span.SetStatus(codes.Ok, "serial compare view success")
// 	return nil
// }

// // CE Conditional expression 条件表达式
// func CE(condition bool, res1 any, res2 any) any {
// 	if condition {
// 		return res1
// 	}
// 	return res2
// }

// func (dss *dataSourceService) updateScanRecord(ctx context.Context, dataSourceId string, taskId string, projectId string) error {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "update scan record")
// 	defer span.End()

// 	record, err := dss.sra.GetByDataSourceIdAndScanner(ctx, dataSourceId, taskId)
// 	if err != nil {
// 		span.SetStatus(codes.Error, "updateScanRecord GetByDataSourceIdAndScanner failed")
// 		logger.Errorf("updateScanRecord GetByDataSourceIdAndScanner Error, %v", err)
// 		return err
// 	}
// 	if len(record) == 0 {
// 		err = dss.sra.CreateScanRecord(ctx, &interfaces.ScanRecord{
// 			RecordID:     xid.New().String(),
// 			DataSourceID: dataSourceId,
// 			Scanner:      CE(taskId == "", interfaces.ManagementScanner, taskId).(string),
// 			ScanTime:     time.Now().UnixMilli(),
// 		})
// 		if err != nil {
// 			span.SetStatus(codes.Error, "updateScanRecord create scan record failed")
// 			logger.Errorf("[updateScanRecord] create scan record rrror, %v", err)
// 			return err
// 		}
// 	} else {
// 		record[0].ScanTime = time.Now().UnixMilli()
// 		err = dss.sra.UpdateScanRecord(ctx, record[0])
// 		if err != nil {
// 			span.SetStatus(codes.Error, "updateScanRecord update scan record failed")
// 			logger.Errorf("[updateScanRecord] update scan record rrror, %v", err)
// 			return err
// 		}
// 	}

// 	if projectId == "" && taskId != "" { //独立任务再增加管理者可见
// 		logger.Infof("updateScanRecord independent task %s", dataSourceId)
// 		err = dss.updateScanRecord(ctx, dataSourceId, "", "")
// 		if err != nil {
// 			span.SetStatus(codes.Error, "updateScanRecord create scan record failed")
// 			return err
// 		}
// 	}

// 	span.SetStatus(codes.Ok, "update scan record success")
// 	return nil
// }

// type FormViewFlag struct {
// 	*interfaces.DataView
// 	flag int
// 	// mu   sync.Mutex
// }
// type FormViewFieldFlag struct {
// 	*interfaces.ViewField
// 	flag int
// }

// func (dss *dataSourceService) createViews(ctx context.Context, dataSource *interfaces.DataSource,
// 	tables []*interfaces.GetDataTableDetailDataBatchRes, dataViewBusinessNameMap map[string]struct{}) (err error) {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "Create data views")
// 	defer span.End()

// 	accountInfo := interfaces.AccountInfo{}
// 	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
// 		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
// 	}

// 	createViews := make([]*interfaces.DataView, 0, len(tables))
// 	for _, table := range tables {
// 		viewId := xid.New().String()
// 		fields := make([]*interfaces.ViewField, len(table.Fields))
// 		var selectFields string
// 		for i, field := range table.Fields {
// 			fields[i] = &interfaces.ViewField{
// 				Name:         field.FieldName,
// 				DisplayName:  dss.AutomaticallyField(ctx, field),
// 				OriginalName: field.FieldName,
// 				Comment:      common.CutStringByCharCount(field.FieldComment, interfaces.CommentCharCountLimit),
// 				Status:       interfaces.ViewScanStatus_New,
// 				PrimaryKey:   sql.NullBool{Bool: field.AdvancedParams.IsPrimaryKey(), Valid: true},
// 				Type:         field.AdvancedParams.GetValue(interfaces.VirtualDataType),
// 				DataLength:   field.FieldLength,
// 				DataAccuracy: field.FieldPrecision,
// 				// OriginalDataType: field.FieldTypeName,
// 				IsNullable: field.AdvancedParams.GetValue(interfaces.IsNullable),
// 			}

// 			if field.AdvancedParams.GetValue(interfaces.VirtualDataType) == "" { //不支持的类型设置状态，跳过创建
// 				fields[i].Status = interfaces.FieldScanStatus_NotSupport
// 			} else {
// 				selectFields = common.CE(selectFields == "", common.QuotationMark(field.FieldName),
// 					fmt.Sprintf("%s,%s", selectFields, common.QuotationMark(field.FieldName))).(string)
// 			}
// 		}

// 		view := &interfaces.DataView{
// 			SimpleDataView: interfaces.SimpleDataView{
// 				ViewID:         viewId,
// 				TechnicalName:  table.Name,
// 				ViewName:       dss.AutomaticallyForm(ctx, table, dataViewBusinessNameMap),
// 				Builtin:        true,
// 				GroupID:        dataSource.ID,
// 				GroupName:      dataSource.Name,
// 				Type:           interfaces.ViewType_Atomic,
// 				QueryType:      interfaces.QueryType_SQL,
// 				DataSourceID:   dataSource.ID,
// 				DataSourceType: dataSource.Type,
// 				Status:         interfaces.ViewScanStatus_New,
// 				Comment:        common.CutStringByCharCount(table.Description, interfaces.CommentCharCountLimit),
// 			},
// 			Fields:         fields,
// 			Creator:        accountInfo,
// 			Updater:        accountInfo,
// 			MetadataFormID: table.Id,
// 			// OriginalName:   table.OrgName,
// 		}

// 		// excel 有catalog, 没有 schema
// 		// tingyun, anyshare7 没有 catalog 和 schema
// 		// 但是这几种类型不支持扫描，理论上走不到这里，可以再判断一次
// 		if dataSource.Type == interfaces.DataSourceType_Excel || dataSource.Type == interfaces.DataSourceType_TingYun ||
// 			dataSource.Type == interfaces.DataSourceType_AS7 {
// 			continue
// 		}

// 		catalogName := dataSource.BinData.CatalogName
// 		schemaName := dataSource.BinData.Schema
// 		// 先用schema，没有再用database
// 		if schemaName == "" {
// 			schemaName = dataSource.BinData.DataBaseName
// 		}

// 		metaTableName := fmt.Sprintf("%s.%s.%s", catalogName, common.QuotationMark(schemaName), common.QuotationMark(table.Name))
// 		// 补齐 sqlstr 和 metatable name
// 		view.SQLStr = fmt.Sprintf("SELECT * FROM %s", metaTableName)
// 		view.MetaTableName = metaTableName
// 		createViews = append(createViews, view)

// 	}
// 	// dataViewSource := dataSource.BinData.DataViewSource

// 	// // 开始事务
// 	// tx, err := dss.db.Begin()
// 	// if err != nil {
// 	// 	span.SetStatus(codes.Error, "CreateView begin DB transaction error")
// 	// 	logger.Errorf("[ScanDataSource] CreateView begin DB transaction error: %s", err.Error())
// 	// 	return rest.NewHTTPError(ctx, http.StatusInternalServerError,
// 	// 		derrors.DataModel_DataView_InternalError_BeginDbTransactionFailed).WithErrorDetails(err.Error())
// 	// }

// 	// needRollback := false
// 	// defer func() {
// 	// 	if !needRollback {
// 	// 		err = tx.Commit()
// 	// 		if err != nil {
// 	// 			errDetails := fmt.Sprintf("[ScanDataSource] CreateView Transaction Commit Failed: %v", err)
// 	// 			span.SetStatus(codes.Error, "CreateView Transaction Commit Failed")
// 	// 			o11y.Error(ctx, errDetails)
// 	// 			logger.Errorf(errDetails)
// 	// 			// // rollback 回滚向vega创建的视图
// 	// 			// if rollbackVegaViewErr := dss.vga.DeleteVegaView(ctx, &interfaces.DeleteViewReq{
// 	// 			// 	CatalogName: dataViewSource,
// 	// 			// 	ViewName:    table.Name,
// 	// 			// }); rollbackVegaViewErr != nil {
// 	// 			// 	errDetails := fmt.Sprintf("[ScanDataSource] DatabaseError and rollback createView by deleteView Error, %v", rollbackVegaViewErr)
// 	// 			// 	span.SetStatus(codes.Error, "CreateView rollback createView by deleteView Error")
// 	// 			// 	o11y.Error(ctx, errDetails)
// 	// 			// 	logger.Errorf(errDetails)
// 	// 			// }
// 	// 		}

// 	// 		logger.Debugf("[ScanDataSource] Transaction Commit Success: %v", formView.TechnicalName)
// 	// 	} else {
// 	// 		err = tx.Rollback()
// 	// 		if err != nil {
// 	// 			errDetails := fmt.Sprintf("[ScanDataSource] Transaction Rollback Error:%v", err)
// 	// 			span.SetStatus(codes.Error, "CreateView Transaction Rollback Error")
// 	// 			o11y.Error(ctx, errDetails)
// 	// 			logger.Errorf(errDetails)
// 	// 		}
// 	// 	}
// 	// }()

// 	logger.Infof("[ScanDataSource] create view count: %d", len(createViews))
// 	if _, err = dss.dvs.CreateDataViews(ctx, createViews, interfaces.ImportMode_Overwrite); err != nil {
// 		errDetails := fmt.Sprintf("[ScanDataSource] create view database error, %v", err)
// 		span.SetStatus(codes.Error, "CreateView create view database error")
// 		o11y.Error(ctx, errDetails)
// 		logger.Errorf(errDetails)
// 		// needRollback = true
// 		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 			WithErrorDetails(errDetails)
// 	}

// 	// // 检查视图在vega是否已经存在
// 	// vegeViews, err := dss.vga.GetVegaViews(ctx, &interfaces.GetViewReq{
// 	// 	CatalogName: catalogName,
// 	// 	SchemaName:  schemaName,
// 	// 	ViewName:    table.Name,
// 	// })
// 	// if err != nil {
// 	// 	needRollback = true
// 	// 	errDetails := fmt.Sprintf("get vega views database error, %v", err)
// 	// 	span.SetStatus(codes.Error, "CreateView get vega views database error")
// 	// 	o11y.Error(ctx, errDetails)
// 	// 	logger.Errorf(errDetails)
// 	// 	return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 	// 		WithErrorDetails(errDetails)
// 	// }

// 	// // 如果没有。就新建
// 	// if len(vegeViews.Entries) == 0 {
// 	// 	// 向vega创建视图
// 	// 	if err = dss.vga.CreateVegaView(ctx, &interfaces.CreateViewReq{
// 	// 		CatalogName: dataViewSource, //虚拟数据源
// 	// 		Query:       createSql,
// 	// 		ViewName:    table.Name,
// 	// 	}); err != nil {
// 	// 		errDetails := fmt.Sprintf("[ScanDataSource] create vega table %s Error, %v, sql=%s", table.Name, err, createSql)
// 	// 		span.SetStatus(codes.Error, "CreateView create vega table Error")
// 	// 		o11y.Warn(ctx, errDetails)
// 	// 		logger.Warnf(errDetails)
// 	// 		createViewErrorCh <- &interfaces.ErrorView{
// 	// 			TechnicalName: table.Name,
// 	// 			Error: rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 	// 				WithErrorDetails(err.Error()),
// 	// 		}
// 	// 		needRollback = true

// 	// 		span.SetStatus(codes.Ok, "ignore and record single vega table error, continue")
// 	// 		return nil
// 	// 	}
// 	// }

// 	span.SetStatus(codes.Ok, "Batch create data views success")
// 	return nil
// }
// func (dss *dataSourceService) AutomaticallyForm(ctx context.Context, table *interfaces.GetDataTableDetailDataBatchRes,
// 	dataViewBusinessNameMap map[string]struct{}) string {
// 	/*
// 		表业务名称按以下顺序自动生成：
// 		    来自加工模型关联的业务表名称
// 		    表注释
// 		    数据理解
// 		    表技术名称
// 	*/

// 	businessName := common.CutStringByCharCount(table.Description, interfaces.BusinessNameCharCountLimit)
// 	if businessName == "" {
// 		businessName = common.CutStringByCharCount(table.Name, interfaces.BusinessNameCharCountLimit)
// 	}

// 	if _, ok := dataViewBusinessNameMap[businessName]; ok {
// 		// 生成2字节随机数
// 		randomBytes := make([]byte, 2)
// 		_, err := rand.Read(randomBytes)
// 		if err != nil {
// 			logger.Errorf("generate random bytes error, %v", err)
// 		}
// 		businessName = common.CutStringByCharCount(fmt.Sprintf("%x_%s", randomBytes, businessName),
// 			interfaces.BusinessNameCharCountLimit)
// 	}

// 	dataViewBusinessNameMap[businessName] = struct{}{}

// 	return businessName
// }

// func (dss *dataSourceService) AutomaticallyField(ctx context.Context, field *interfaces.FieldsBatch) (businessName string) {
// 	/*
// 		列业务名称按以下顺序自动生成：
// 		    来自加工模型关联的业务表“字段中文名称
// 		    字段注释
// 		    数据理解
// 		    列技术名称
// 	*/
// 	if businessName == "" {
// 		businessName = common.CutStringByCharCount(field.FieldComment, interfaces.BusinessNameCharCountLimit)
// 	}
// 	if businessName == "" {
// 		businessName = common.CutStringByCharCount(field.FieldName, interfaces.BusinessNameCharCountLimit)
// 	}
// 	return
// }

// // 更新 FormView
// func (dss *dataSourceService) updateView(ctx context.Context, dataSource *interfaces.DataSource, view *interfaces.DataView,
// 	table *interfaces.GetDataTableDetailDataBatchRes) (err error) {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "update one view")
// 	defer span.End()

// 	// 获取字段列表
// 	views, err := dss.dvs.GetDataViews(ctx, []string{view.ViewID}, false)
// 	if err != nil {
// 		span.SetStatus(codes.Error, "Get data views failed")
// 		return err
// 	}

// 	if len(views) == 0 {
// 		span.SetStatus(codes.Error, "Data view not found")
// 		return rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataView_DataViewNotFound)
// 	}
// 	fieldList := views[0].Fields
// 	logger.Debugf("update view, table name: %s, view name: %s", table.Name, view.ViewName)
// 	logger.Debugf("view fields count is %d, view fields list is %+v", len(fieldList), fieldList)
// 	logger.Debugf("metadata tables fields count is %d, fields list is %+v", len(table.Fields), table.Fields)

// 	newFields := make([]*interfaces.ViewField, 0)
// 	updateFields := make([]*interfaces.ViewField, 0)
// 	deleteFields := make([]string, 0)

// 	// 已有的字段列表
// 	fieldsMap := make(map[string]*FormViewFieldFlag)
// 	for _, field := range fieldList {
// 		fieldsMap[field.OriginalName] = &FormViewFieldFlag{ViewField: field, flag: 1}
// 	}
// 	formViewModify := false

// 	var selectFields string
// 	final_view_fields := make([]*interfaces.ViewField, 0, len(table.Fields))
// 	for _, field := range table.Fields {
// 		if oldField, ok := fieldsMap[field.FieldName]; !ok {
// 			logger.Debugf("update view, table name: %s, field name: %s, field not exist in view", table.Name, field.FieldName)
// 			//field new
// 			logger.Debugf("update view, table name: %s, field name: %s, field not exist in view", table.Name, field.FieldName)
// 			newField := &interfaces.ViewField{
// 				Name:         field.FieldName,
// 				DisplayName:  dss.AutomaticallyField(ctx, field),
// 				OriginalName: field.FieldName,
// 				Comment:      common.CutStringByCharCount(field.FieldComment, interfaces.CommentCharCountLimit),
// 				Status:       interfaces.FieldScanStatus_New,
// 				PrimaryKey:   sql.NullBool{Bool: field.AdvancedParams.IsPrimaryKey(), Valid: true},
// 				Type:         field.AdvancedParams.GetValue(interfaces.VirtualDataType),
// 				DataLength:   field.FieldLength,
// 				DataAccuracy: field.FieldPrecision,
// 				IsNullable:   field.AdvancedParams.GetValue(interfaces.IsNullable),
// 			}
// 			newFields = append(newFields, newField)
// 			final_view_fields = append(final_view_fields, newField)
// 			formViewModify = true

// 			if newField.Type == "" { //不支持的类型设置状态，跳过创建
// 				newField.Status = interfaces.FieldScanStatus_NotSupport
// 			} else {
// 				selectFields = common.CE(selectFields == "", common.QuotationMark(field.FieldName),
// 					fmt.Sprintf("%s,%s", selectFields, common.QuotationMark(field.FieldName))).(string)
// 			}
// 		} else {
// 			// field update
// 			logger.Debugf("update view, table name: %s, field name: %s, field type change from %s to %s", table.Name, field.FieldName, oldField.Type, field.AdvancedParams.GetValue(interfaces.VirtualDataType))
// 			originalDataTypeChange := dss.originalDataTypeChange(oldField, field)
// 			switch {
// 			case originalDataTypeChange: //字段类型变更
// 				//field  VirtualDataType  update
// 				modified_field := dss.updateFieldStruct(oldField, field)
// 				updateFields = append(updateFields, modified_field)
// 				final_view_fields = append(final_view_fields, modified_field)
// 				formViewModify = true
// 			case !originalDataTypeChange && oldField.Status == interfaces.FieldScanStatus_Delete: //删除的反转为新增
// 				logger.Infof("FormViewFieldDelete status Reversal, oldField name = %s", oldField.OriginalName)
// 				comment := common.CutStringByCharCount(field.FieldComment, interfaces.CommentCharCountLimit)
// 				modified_field := &interfaces.ViewField{
// 					Name:              oldField.Name,
// 					Type:              oldField.Type,
// 					Comment:           comment,
// 					DisplayName:       oldField.DisplayName,
// 					OriginalName:      oldField.OriginalName,
// 					DataLength:        oldField.DataLength,
// 					DataAccuracy:      oldField.DataAccuracy,
// 					Status:            interfaces.FieldScanStatus_New,
// 					IsNullable:        oldField.IsNullable,
// 					BusinessTimestamp: oldField.BusinessTimestamp,
// 					PrimaryKey:        oldField.PrimaryKey,
// 				}
// 				updateFields = append(updateFields, modified_field)
// 				final_view_fields = append(final_view_fields, modified_field)
// 				formViewModify = true
// 			case oldField.Comment != field.FieldComment: // 只修改字段备注视作不变状态
// 				oldField.ViewField.Comment = common.CutStringByCharCount(field.FieldComment, interfaces.CommentCharCountLimit)
// 				updateFields = append(updateFields, oldField.ViewField)
// 				final_view_fields = append(final_view_fields, oldField.ViewField)
// 			default: //field not update
// 				final_view_fields = append(final_view_fields, oldField.ViewField)
// 			}

// 			fieldsMap[field.FieldName].flag = 2

// 			// newDataType := field.AdvancedParams.GetValue(interfaces.VirtualDataType)
// 			// selectField := dss.genSelectSQL(originalDataTypeChange, newDataType, oldField, &updateFields)
// 			selectField := oldField.OriginalName
// 			if originalDataTypeChange && field.AdvancedParams.GetValue(interfaces.VirtualDataType) == "" { //不支持的类型设置状态，跳过创建
// 				updateFields[len(updateFields)-1].Status = interfaces.FieldScanStatus_NotSupport
// 			} else {
// 				selectFields = common.CE(selectFields == "", common.QuotationMark(selectField),
// 					fmt.Sprintf("%s,%s", selectFields, common.QuotationMark(selectField))).(string)
// 			}
// 		}
// 	}

// 	// 删除的字段先不添加到final_view_fields，因为还没与AF对接，不需要标记字段状态
// 	for _, field := range fieldsMap {
// 		if field.flag == 1 {
// 			//field delete
// 			deleteFields = append(deleteFields, field.OriginalName)
// 			formViewModify = true
// 		}
// 	}

// 	// 更新视图的字段列表
// 	view.Fields = final_view_fields

// 	// formViewUpdate := view.Comment != table.Description || view.OriginalName != table.OrgName
// 	formViewUpdate := view.Comment != table.Description
// 	if formViewUpdate {
// 		view.Comment = common.CutStringByCharCount(table.Description, interfaces.CommentCharCountLimit)
// 		// view.OriginalName = table.OrgName
// 	}

// 	var query string
// 	typeName := dataSource.Type
// 	schemaName := dataSource.BinData.Schema
// 	// 下面几种类型没有schema, 不过这几种类型也不支持扫描
// 	if typeName != interfaces.DataSourceType_Excel && typeName != interfaces.DataSourceType_TingYun &&
// 		typeName != interfaces.DataSourceType_AS7 {
// 		if schemaName == "" {
// 			schemaName = dataSource.BinData.DataBaseName
// 		}
// 	}

// 	// dataViewSource := dataSource.BinData.DataViewSource
// 	query = fmt.Sprintf("SELECT * FROM %s.%s.%s", dataSource.BinData.CatalogName,
// 		common.QuotationMark(schemaName), common.QuotationMark(table.Name))

// 	if formViewModify { //表的字段有变化

// 		// if err = dss.vga.ModifyVegaView(ctx, &interfaces.ModifyViewReq{
// 		// 	CatalogName: dataViewSource,
// 		// 	Query:       query,
// 		// 	ViewName:    table.Name,
// 		// }); err != nil {
// 		// 	span.SetStatus(codes.Error, "Modify vega view failed")
// 		// 	o11y.Error(ctx, fmt.Sprintf("Modify vega view error, %v", err))
// 		// 	logger.Errorf("Modify vega view DatabaseError failed, %v, sql=%s", err, query)

// 		// 	createViewErrorCh <- &interfaces.ErrorView{
// 		// 		Id:            view.ViewID,
// 		// 		TechnicalName: table.Name,
// 		// 		Error: rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_InternalError_DataBaseError).
// 		// 			WithErrorDetails(err.Error()),
// 		// 	}

// 		// 	span.SetStatus(codes.Ok, "Modify vega view error, ignore and record and continue")
// 		// 	return nil
// 		// }
// 		logger.Infof("Modify vega view, table name=%s, formView ID=%s", table.Name, view.ViewID)

// 		if view.Status == interfaces.ViewScanStatus_NoChange || view.Status == interfaces.ViewScanStatus_New {
// 			view.Status = interfaces.ViewScanStatus_Modify
// 			formViewUpdate = true
// 		}
// 	} else { //表的字段无变化
// 		if view.Status == interfaces.ViewScanStatus_New || view.Status == interfaces.ViewScanStatus_Modify {
// 			view.Status = interfaces.ViewScanStatus_NoChange //二次扫描无变化 视图状态变为无变化
// 			formViewUpdate = true
// 		}
// 	}

// 	if view.Status == interfaces.ViewScanStatus_Delete { //删除状态又找到
// 		logger.Infof("FormViewDelete status Reversal, formView ID=%s", view.ViewID)
// 		view.Status = interfaces.ViewScanStatus_New //删除状态表反转为新建
// 		formViewUpdate = true
// 	}

// 	accountInfo := interfaces.AccountInfo{}
// 	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
// 		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
// 	}

// 	if len(newFields) != 0 || len(updateFields) != 0 || len(deleteFields) != 0 { //字段及表都修改
// 		view.Updater = accountInfo
// 		view.UpdateTime = time.Now().UnixMilli()
// 		view.Status = interfaces.ViewScanStatus_Modify
// 		view.SQLStr = query

// 		if err = dss.dvs.UpdateDataView(ctx, nil, view); err != nil {
// 			span.SetStatus(codes.Error, "Update data view failed")
// 			o11y.Error(ctx, fmt.Sprintf("Update data view error, %v", err))
// 			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
// 				derrors.DataModel_InternalError_DataBaseError).WithErrorDetails(err.Error())
// 		}

// 		// } else if formViewUpdate || newUniformCatalogCode { //只反转，字段不变更，或分配新的 UniformCatalogCode
// 	} else if formViewUpdate { //只反转，字段不变更，或分配新的 UniformCatalogCode
// 		view.Updater = accountInfo
// 		view.UpdateTime = time.Now().UnixMilli()
// 		view.SQLStr = query

// 		if err = dss.dvs.UpdateDataView(ctx, nil, view); err != nil {
// 			span.SetStatus(codes.Error, "Update data view failed")
// 			o11y.Error(ctx, fmt.Sprintf("Update data view error, %v", err))
// 			logger.Errorf("update view database failed, %v", err)
// 			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
// 				derrors.DataModel_InternalError_DataBaseError).WithErrorDetails(err.Error())
// 		}
// 	}

// 	span.SetStatus(codes.Ok, "Update single data view success")
// 	return nil
// }

// func (dss *dataSourceService) originalDataTypeChange(oldField *FormViewFieldFlag, field *interfaces.FieldsBatch) bool {
// 	if oldField.OriginalName != field.FieldName {
// 		return true
// 	}

// 	if oldField.Type != field.AdvancedParams.GetValue(interfaces.VirtualDataType) {
// 		return true
// 	}

// 	if oldField.IsNullable != field.AdvancedParams.GetValue(interfaces.IsNullable) {
// 		return true
// 	}

// 	if oldField.DataAccuracy != field.FieldPrecision {
// 		return true
// 	}

// 	if oldField.DataLength != field.FieldLength {
// 		return true
// 	}

// 	if oldField.PrimaryKey.Bool != field.AdvancedParams.IsPrimaryKey() {
// 		return true
// 	}

// 	//虚拟化数据类型 未处理
// 	//是否为空、comment、COLUMN_DEF、IS_NULLABLE，原因：不能显示
// 	return false
// }

// // func (dss *dataSourceService) genSelectSQL(originalDataTypeChange bool, scanNewDataType string, oldField *FormViewFieldFlag, updateFields *[]*interfaces.ViewField) string {
// // 	originalName := common.QuotationMark(oldField.OriginalName)
// // 	if originalDataTypeChange { //字段类型变更
// // 		var updateField *interfaces.ViewField
// // 		updateFieldsTmp := *updateFields
// // 		if len(updateFieldsTmp) > 0 {
// // 			updateField = updateFieldsTmp[len(*updateFields)-1]
// // 		}
// // 		if _, exist := dtype.TypeConvertMap[scanNewDataType+oldField.Type]; exist { //扫描新类型转为原来重置类型
// // 			var selectField string
// // 			switch oldField.Type { //扫描转换
// // 			case dtype.DATE, dtype.TIME, dtype.TIME_WITH_TIME_ZONE, dtype.DATETIME, dtype.TIMESTAMP, dtype.TIMESTAMP_WITH_TIME_ZONE:
// // 				//扫描预设类型是scanNewDataType 不是beforeDataType:= util.CE(field.ResetBeforeDataType.String != "", field.ResetBeforeDataType.String, field.DataType).(string)
// // 				if (scanNewDataType == dtype.CHAR || scanNewDataType == dtype.VARCHAR || scanNewDataType == dtype.STRING) && oldField.ResetConvertRules.String != "" {
// // 					selectField = fmt.Sprintf("try_cast(date_parse(%s,'%s') AS %s) %s", originalName, oldField.ResetConvertRules.String, scanNewDataType, originalName)
// // 				} else {
// // 					selectField = fmt.Sprintf("try_cast(%s AS %s) %s", originalName, scanNewDataType, originalName)
// // 				}
// // 			case dtype.DECIMAL, dtype.NUMERIC, dtype.DEC:
// // 				selectField = fmt.Sprintf("try_cast(%s AS %s(%d,%d)) %s", originalName, scanNewDataType, oldField.DataLength, oldField.DataAccuracy, originalName)
// // 			default:
// // 				selectField = fmt.Sprintf("try_cast(%s AS %s) %s", originalName, scanNewDataType, originalName)
// // 			}
// // 			updateField.Type = oldField.Type //当前数据类型改为非预设类型，重新扫描，预设类型属性值会自动更新，但不影响已选的类型（如：当前数据类型选项为A，预设类型为B，此时重新扫描预设类型变为C，当前类型选项仍然为A，但预设值会从B更新为C）
// // 			updateField.ResetBeforeDataType = sql.NullString{String: scanNewDataType, Valid: true}
// // 			return selectField
// // 		} else { //其他情况改为预设
// // 			updateField.Type = scanNewDataType
// // 		}
// // 	}
// // 	if !originalDataTypeChange { //保持原有类型转换
// // 		var selectField string
// // 		switch oldField.Type {
// // 		case dtype.DATE, dtype.TIME, dtype.TIME_WITH_TIME_ZONE, dtype.DATETIME, dtype.TIMESTAMP, dtype.TIMESTAMP_WITH_TIME_ZONE:
// // 			beforeDataType := common.CE(oldField.ResetBeforeDataType.String != "", oldField.ResetBeforeDataType.String, oldField.Type).(string)
// // 			if beforeDataType == dtype.CHAR || beforeDataType == dtype.VARCHAR || beforeDataType == dtype.STRING {
// // 				selectField = fmt.Sprintf("try_cast(date_parse(%s,'%s') AS %s) %s", originalName, oldField.ResetConvertRules.String, scanNewDataType, originalName)
// // 			} else {
// // 				selectField = fmt.Sprintf("try_cast(%s AS %s) %s", originalName, scanNewDataType, originalName)
// // 			}
// // 		case dtype.DECIMAL, dtype.NUMERIC, dtype.DEC:
// // 			selectField = fmt.Sprintf("try_cast(%s AS %s(%d,%d)) %s", originalName, scanNewDataType, oldField.DataLength, oldField.DataAccuracy, originalName)
// // 		default:
// // 			selectField = fmt.Sprintf("try_cast(%s AS %s) %s", originalName, scanNewDataType, originalName)
// // 			return selectField
// // 		}
// // 	}
// // 	return originalName
// // }

// func (dss *dataSourceService) updateFieldStruct(oldField *FormViewFieldFlag, field *interfaces.FieldsBatch) *interfaces.ViewField {
// 	updateField := &interfaces.ViewField{}
// 	if err := copier.Copy(updateField, oldField); err != nil {
// 		logger.Error("updateFieldStruct  copier.Copy err", zap.Error(err))
// 	}

// 	if oldField.Status != interfaces.FieldScanStatus_New { //新建的修改还是新建
// 		updateField.Status = interfaces.FieldScanStatus_Modify
// 	}

// 	updateField.PrimaryKey = sql.NullBool{Bool: field.AdvancedParams.IsPrimaryKey(), Valid: true}
// 	updateField.Type = field.AdvancedParams.GetValue(interfaces.VirtualDataType)
// 	updateField.DataLength = field.FieldLength
// 	updateField.DataAccuracy = field.FieldPrecision
// 	updateField.IsNullable = field.AdvancedParams.GetValue(interfaces.IsNullable)
// 	updateField.Comment = common.CutStringByCharCount(field.FieldComment, interfaces.CommentCharCountLimit)
// 	updateField.OriginalName = field.FieldName
// 	// 扫描字段名和技术名称保持一致
// 	updateField.Name = field.FieldName

// 	return updateField
// }

// func (dss *dataSourceService) MarkViewAsSourceDeleted(ctx context.Context, viewsIDs []string) error {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Mark view as source deleted")
// 	defer span.End()

// 	logger.Infof("MultipleScan deleteView %+v", viewsIDs)
// 	if len(viewsIDs) == 0 {
// 		span.SetStatus(codes.Ok, "viewsIDs is empty")
// 		return nil
// 	}

// 	param := &interfaces.UpdateViewStatus{
// 		ViewStatus: interfaces.ViewScanStatus_Delete,
// 		UpdateTime: time.Now().UnixMilli(),
// 	}

// 	//  更新视图状态为已删除
// 	if err := dss.dvs.UpdateViewStatus(ctx, viewsIDs, param); err != nil {
// 		span.SetStatus(codes.Error, "update view status failed")
// 		o11y.Error(ctx, "update view status failed")
// 		return err
// 	}

// 	span.SetStatus(codes.Ok, "update view status success")
// 	return nil
// }

// func (dss *dataSourceService) FinishProject(ctx context.Context, req *interfaces.FinishProjectReq) error {
// 	records, err := dss.sra.GetByTaskIds(ctx, req.TaskIDs)
// 	if err != nil {
// 		logger.Errorf("FinishProject GetByScanner Error, %v", err)
// 		return err
// 	}
// 	for _, record := range records {
// 		manageRecord, err := dss.sra.GetByDataSourceIdAndScanner(ctx, record.DataSourceID, interfaces.ManagementScanner)
// 		if err != nil {
// 			logger.Errorf("FinishProject GetByDatasourceIdAndScanner Error, %v", err)
// 			return err
// 		}
// 		if len(manageRecord) == 0 {
// 			err = dss.sra.CreateScanRecord(ctx, &interfaces.ScanRecord{
// 				RecordID:     xid.New().String(),
// 				DataSourceID: record.DataSourceID,
// 				Scanner:      interfaces.ManagementScanner,
// 				ScanTime:     time.Now().UnixMilli(),
// 			})
// 			if err != nil {
// 				logger.Errorf("FinishProject Create Error, %v", err)
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// // func (dss *dataSourceService) ScanIndexBase(ctx context.Context, req *interfaces.ScanTask) (*interfaces.ScanResult, error) {
// // 	// 获取索引库列表
// // 	bases, err := dss.iba.ListIndexBases(ctx)
// // 	if err != nil {
// // 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// // 			WithErrorDetails(fmt.Sprintf("list index bases failed, %v", err))
// // 	}

// // 	// 获取index_base 分组下的所有元数据视图
// // 	views, _, err := dss.dvs.ListDataViews(ctx, &interfaces.ListViewQueryParams{
// // 		GroupID:   "__index_base",
// // 		GroupName: "index_base",
// // 	})
// // 	if err != nil {
// // 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// // 			WithErrorDetails(fmt.Sprintf("list data views failed, %v", err))
// // 	}

// // 	// 遍历 bases，和 views对比，得出需要创建、更新、删除的视图列表
// // 	viewsMap := make(map[string]*interfaces.SimpleDataView)
// // 	for _, view := range views {
// // 		viewsMap[view.ViewID] = view
// // 	}

// // 	basesMap := make(map[string]interfaces.SimpleIndexBase)
// // 	basesList := make([]string, 0, len(bases))
// // 	for _, base := range bases {
// // 		basesMap[base.BaseType] = base
// // 		basesList = append(basesList, base.BaseType)
// // 	}

// // 	// 获取索引库的字段详情
// // 	detailedBases, err := dss.iba.GetManyIndexBasesByTypes(ctx, basesList)
// // 	if err != nil {
// // 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// // 			WithErrorDetails(fmt.Sprintf("get many index bases by types failed, %v", err))
// // 	}

// // 	viewsForCreate := []*interfaces.DataView{}
// // 	viewsForUpdate := []*interfaces.DataView{}
// // 	viewIDsForDelete := []string{}

// // 	// 为每个索引库创建对应的元数据视图
// // 	for _, base := range detailedBases {
// // 		viewID := generateDataViewID(base.BaseType)
// // 		if view, ok := viewsMap[viewID]; ok {
// // 			// 获取索引库的字段
// // 			allBaseFields := mergeIndexBaseFields2Map(base.Mappings)

// // 			// 获取视图详情
// // 			viewDetail, err := dss.dvs.GetDataView(ctx, viewID)
// // 			if err != nil {
// // 				return nil, err
// // 			}

// // 			// 将视图的字段转成map
// // 			viewFieldsMap := make(map[string]*interfaces.ViewField)
// // 			for _, vField := range viewDetail.Fields {
// // 				viewFieldsMap[vField.OriginalName] = vField
// // 			}

// // 			fieldChange := false
// // 			newViewFields := copyFields(viewDetail.Fields)
// // 			for _, baseField := range allBaseFields {
// // 				if _, ok := viewFieldsMap[baseField.Field]; !ok {
// // 					// 索引库的字段在视图中不存在，需要添加
// // 					fieldChange = true
// // 					newViewFields = append(newViewFields, &interfaces.ViewField{
// // 						Name:         baseField.Field,
// // 						Type:         baseField.Type,
// // 						Comment:      "",
// // 						DisplayName:  baseField.DisplayName,
// // 						OriginalName: baseField.Field,
// // 						Status:       interfaces.FieldScanStatus_New,
// // 						// DataLength:        0,
// // 						// DataAccuracy:      0,
// // 						// IsNullable:       "",
// // 						// BusinessTimestamp: false,
// // 						// PrimaryKey:       sql.NullBool{},
// // 					})
// // 				} else {
// // 					// 索引库的字段在视图中存在，需要检查字段是否变更
// // 					vField := viewFieldsMap[baseField.Field]
// // 					if isIndexBaseFieldChange(baseField, vField) {
// // 						// 字段变更，需要更新
// // 						fieldChange = true
// // 						newViewFields = append(newViewFields, &interfaces.ViewField{
// // 							Name:         vField.Name,
// // 							Type:         vField.Type,
// // 							Comment:      "",
// // 							DisplayName:  vField.DisplayName,
// // 							OriginalName: vField.OriginalName,
// // 							Status:       interfaces.FieldScanStatus_Modify,
// // 							// DataLength:        0,
// // 							// DataAccuracy:      0,
// // 							// IsNullable:       "",
// // 							// BusinessTimestamp: false,
// // 							// PrimaryKey:       sql.NullBool{},
// // 						})
// // 					}
// // 				}
// // 			}

// // 			for _, vField := range viewFieldsMap {
// // 				if _, ok := allBaseFields[vField.OriginalName]; !ok {
// // 					// 视图的字段在索引库中不存在，需要删除
// // 					fieldChange = true
// // 					newViewFields = append(newViewFields, &interfaces.ViewField{
// // 						Name:         vField.Name,
// // 						Type:         vField.Type,
// // 						Comment:      "",
// // 						DisplayName:  vField.DisplayName,
// // 						OriginalName: vField.OriginalName,
// // 						Status:       interfaces.FieldScanStatus_Delete,
// // 						// DataLength:        0,
// // 						// DataAccuracy:      0,
// // 						// IsNullable:       "",
// // 						// BusinessTimestamp: false,
// // 						// PrimaryKey:       sql.NullBool{},
// // 					})
// // 				}
// // 			}

// // 			// 判断字段列表是否变化
// // 			if fieldChange {
// // 				// 字段列表变化，需要更新视图
// // 				viewsForUpdate = append(viewsForUpdate, &interfaces.DataView{
// // 					SimpleDataView: interfaces.SimpleDataView{
// // 						ViewID:        viewID,
// // 						ViewName:      view.ViewName,
// // 						TechnicalName: view.TechnicalName,
// // 						GroupID:       view.GroupID,
// // 						GroupName:     view.GroupName,
// // 						Type:          interfaces.ViewType_Atomic,
// // 						QueryType:     interfaces.QueryType_DSL,
// // 						Status:        interfaces.ViewScanStatus_Modify,
// // 					},
// // 				})
// // 			}
// // 		} else {
// // 			// 创建视图
// // 			viewsForCreate = append(viewsForCreate, &interfaces.DataView{
// // 				SimpleDataView: interfaces.SimpleDataView{
// // 					ViewID:        viewID,
// // 					TechnicalName: generateDataViewName(base.BaseType),
// // 					ViewName:      generateDataViewName(base.BaseType),
// // 					GroupID:       "__index_base",
// // 					GroupName:     "index_base",
// // 					Type:          interfaces.ViewType_Atomic,
// // 					QueryType:     interfaces.QueryType_DSL,
// // 					Status:        interfaces.ViewScanStatus_New,
// // 				},
// // 			})
// // 		}
// // 	}

// // 	// 对比找出需要删除的视图列表
// // 	for _, view := range views {
// // 		if _, ok := basesMap[view.ViewName]; !ok {
// // 			viewIDsForDelete = append(viewIDsForDelete, view.ViewID)
// // 		}
// // 	}

// // 	// 创建和更新视图
// // 	if _, err := dss.dvs.CreateDataViews(ctx, nil, viewsForCreate, "", interfaces.ImportMode_Normal); err != nil {
// // 		return nil, err
// // 	}

// // 	// 循环调用更新函数，其中校验对象的更新权限，如果没有更新权限则报错
// // 	for _, uView := range viewsForUpdate {
// // 		err = dss.dvs.UpdateDataView(ctx, nil, uView)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 	}

// // 	// 删除视图
// // 	if err := dss.dvs.DeleteDataViews(ctx, viewIDsForDelete); err != nil {
// // 		return nil, err
// // 	}

// // 	return &interfaces.ScanResult{}, nil
// // }

// // func (dss *dataSourceService) ConcurrentScan(ctx context.Context, req *interfaces.ScanTask) (*interfaces.ScanResult, error) {
// // 	res := &interfaces.ScanResult{}
// // 	//获取数据源信息
// // 	dataSource, err := dss.dsa.GetDataSourceByID(ctx, req.DataSourceID)
// // 	if err != nil {
// // 		return res, err
// // 	}

// // 	if dataSource.Type == interfaces.DataSourceType_Excel {
// // 		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
// // 			WithErrorDetails("Excel-type data sources do not support scanning")
// // 	}

// // 	// data_source,status从scan_record表获取
// // 	record, exist, err := dss.sra.GetByDataSourceId(ctx, dataSource.ID)
// // 	if err != nil {
// // 		logger.Errorf("get scan record Error, %v", err)
// // 		return nil, err
// // 	}

// // 	if !exist {
// // 		err = dss.sra.CreateScanRecord(ctx, &interfaces.ScanRecord{
// // 			RecordID:         xid.New().String(),
// // 			DataSourceID:     dataSource.ID,
// // 			Scanner:          interfaces.ManagementScanner,
// // 			ScanTime:         time.Now().UnixMilli(),
// // 			DataSourceStatus: interfaces.DataSourceScanning,
// // 		})
// // 		if err != nil {
// // 			logger.Errorf("updateScanRecord Create Error", zap.Error(err))
// // 			return res, err
// // 		}
// // 	} else if record.DataSourceStatus == interfaces.DataSourceScanning {
// // 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// // 			WithErrorDetails("data source is scanning already")
// // 	}

// // 	//设置数据源扫描中
// // 	record.DataSourceStatus = interfaces.DataSourceScanning
// // 	if err = dss.sra.UpdateScanRecordStatus(ctx, &interfaces.ScanRecordStatus{
// // 		ID:     record.RecordID,
// // 		Status: interfaces.DataSourceScanning,
// // 	}); err != nil {
// // 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// // 			WithErrorDetails(fmt.Sprintf("update data source status failed, %v", err))
// // 	}
// // 	defer func() {
// // 		//设置数据源可用
// // 		record.DataSourceStatus = interfaces.DataSourceAvailable
// // 		if err = dss.sra.UpdateScanRecordStatus(ctx, &interfaces.ScanRecordStatus{
// // 			ID:     record.RecordID,
// // 			Status: interfaces.DataSourceAvailable,
// // 		}); err != nil {
// // 			logger.Errorf("[ScanDataSource] UpdateDataSource DataSourceAvailable Error, %v", err)
// // 		}
// // 	}()

// // 	//采集元数据
// // 	if err = dss.Collect(ctx, dataSource); err != nil {
// // 		logger.Errorf("[ScanDataSource] Scan Collect Error, %v", err)
// // 		return res, err
// // 	}

// // 	// dataSource
// // 	// err, _, _, result := getInfoFromBinData(dataSource)
// // 	// if err != nil {
// // 	// 	logger.Errorf("make json failed from bin_data: %s", string(dataSource.BinData))
// // 	// 	return res, err
// // 	// }
// // 	// dataViewSource, ok := result["data_view_source"]
// // 	// if !ok || dataViewSource == nil {
// // 	// 	// data_view_source 一定会存在：dip新的data_source版本中
// // 	// 	logger.Errorf("data_view_source not found from bin_data: %s", string(dataSource.BinData))
// // 	// 	return res, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// // 	// 		WithErrorDetails("data_view_source not found from bin_data")
// // 	// }
// // 	// 由于data_view_source 一定会存在：dip新的data_source版本中，所以不需要手动再创建了
// // 	//创建虚拟视图
// // 	// if dataSource.DataViewSource == "" {
// // 	// 	if err = dss.genDataViewSource(ctx, dataSource); err != nil {
// // 	// 		return res, err
// // 	// 	}
// // 	// }

// // 	//对比采集元数据
// // 	err = dss.GetMetadataAndCompare(ctx, req, dataSource, res)
// // 	if err != nil {
// // 		return res, err
// // 	}

// // 	return res, nil
// // }

// // // 获取元数据并比较
// // func (dss *dataSourceService) GetMetadataAndCompare(ctx context.Context, req *interfaces.ScanTask,
// // 	dataSource *interfaces.DataSource, res *interfaces.ScanResult) error {
// // 	formViews, err := dss.dvs.GetDataViewsBySourceID(ctx, dataSource.ID)
// // 	if err != nil {
// // 		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// // 			WithErrorDetails(err.Error())
// // 	}
// // 	formViewsMap := make(map[string]*FormViewFlag)
// // 	// 维护业务名称map，避免新生成的业务名称会在分组内重复
// // 	dataViewBusinessNameMap := make(map[string]struct{})
// // 	for _, formView := range formViews {
// // 		formViewsMap[formView.TechnicalName] = &FormViewFlag{DataView: formView, flag: 1}
// // 		dataViewBusinessNameMap[formView.ViewName] = struct{}{}
// // 	}

// // 	allTable, err := dss.GetDataSourceAllTableInfo(ctx, dataSource)
// // 	if err != nil {
// // 		logger.Errorf("[ScanDataSource] Scan Collect Error, %v", err)
// // 		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// // 			WithErrorDetails(err.Error())
// // 	}

// // 	tablesMap := make(map[string]*interfaces.GetDataTableDetailDataBatchRes)
// // 	for _, table := range allTable {
// // 		tablesMap[table.Name] = table
// // 	}

// // 	// delete view
// // 	deleteViewIDs := make([]string, 0)
// // 	for techName, vv := range formViewsMap {
// // 		if _, ok := tablesMap[techName]; !ok {
// // 			// 源表被删除了，标记为源表删除
// // 			deleteViewIDs = append(deleteViewIDs, vv.ViewID)

// // 		}
// // 	}

// // 	err = dss.MarkViewAsSourceDeleted(ctx, deleteViewIDs)
// // 	if err != nil {
// // 		return err
// // 	}

// // 	dataViewSource := dataSource.BinData.DataViewSource

// // 	logger.Infof("[vma] get table count %d ", len(allTable))
// // 	if len(allTable) == 0 || dataViewSource == "" {
// // 		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// // 			WithErrorDetails("data source has no table")
// // 	}

// // 	//增加扫描记录
// // 	if err = dss.updateScanRecord(ctx, dataSource.ID, req.TaskID, req.ProjectID); err != nil {
// // 		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// // 			WithErrorDetails(fmt.Sprintf("update scan record failed, %v", err))
// // 	}

// // 	allCount := len(allTable)
// // 	wg := &sync.WaitGroup{}
// // 	//统计失败视图
// // 	createViewErrorCh := make(chan *interfaces.ErrorView)
// // 	createViewError := make([]*interfaces.ErrorView, 0)

// // 	go dss.Receive(ctx, createViewErrorCh, &createViewError)

// // 	//计算ve耗时
// // 	costCh := make(chan *VETimeCost)
// // 	cost := &VETimeCost{}
// // 	go dss.CostCalculate(costCh, cost)

// // 	if allCount/interfaces.GoroutineMinTableCount >= interfaces.ConcurrentCount { //最大并发度
// // 		wg.Add(interfaces.ConcurrentCount)
// // 		singleGoroutineDealTableCount := allCount / interfaces.ConcurrentCount
// // 		for i := 0; i < interfaces.ConcurrentCount; i++ {
// // 			start := i * singleGoroutineDealTableCount
// // 			end := (i + 1) * singleGoroutineDealTableCount                                                                         //end := (i+1)*singleGoroutineDealTableCount -1 不包含end
// // 			go dss.CompareView(ctx, wg, createViewErrorCh, dataSource, formViewsMap, allTable[start:end], dataViewBusinessNameMap) //dataSource 只读不写; formViewsMap;
// // 		}
// // 		doneTable := singleGoroutineDealTableCount * interfaces.ConcurrentCount
// // 		if remain := allCount - doneTable; remain > 0 {
// // 			dss.CompareView(ctx, nil, createViewErrorCh, dataSource, formViewsMap, allTable[doneTable:], dataViewBusinessNameMap)
// // 		}

// // 	} else if allCount > interfaces.GoroutineMinTableCount && allCount/interfaces.GoroutineMinTableCount <= interfaces.ConcurrentCount { //适时并发度
// // 		count := allCount / interfaces.GoroutineMinTableCount
// // 		wg.Add(count)
// // 		singleGoroutineDealTableCount := allCount / count
// // 		for i := 0; i < count; i++ {
// // 			start := i * singleGoroutineDealTableCount
// // 			end := (i + 1) * singleGoroutineDealTableCount
// // 			go dss.CompareView(ctx, wg, createViewErrorCh, dataSource, formViewsMap, allTable[start:end], dataViewBusinessNameMap)
// // 		}
// // 		doneTable := singleGoroutineDealTableCount * count
// // 		if remain := allCount - doneTable; remain > 0 {
// // 			dss.CompareView(ctx, nil, createViewErrorCh, dataSource, formViewsMap, allTable[doneTable:], dataViewBusinessNameMap)
// // 		}

// // 	} else if allCount <= interfaces.GoroutineMinTableCount { //零并发度
// // 		dss.CompareView(ctx, nil, createViewErrorCh, dataSource, formViewsMap, allTable, dataViewBusinessNameMap)
// // 	}

// // 	wg.Wait()
// // 	close(createViewErrorCh)
// // 	close(costCh)
// // 	res.ErrorView = createViewError
// // 	res.ErrorViewCount = len(createViewError)
// // 	res.ScanViewCount = len(allTable)

// // 	logger.Infof("createViewCost time %d ,createViewCount %d, createViewMax time %d",
// // 		cost.createViewCost.Milliseconds(), cost.createViewCount, cost.createViewMax.Milliseconds())
// // 	logger.Infof("updateViewCost time %d ,updateViewCount %d, updateViewMax time %d",
// // 		cost.updateViewCost.Milliseconds(), cost.updateViewCount, cost.updateViewMax.Milliseconds())
// // 	return nil
// // }

// // func (dss *dataSourceService) CompareView(ctx context.Context, wg *sync.WaitGroup, createViewErrorCh chan *interfaces.ErrorView,
// // 	dataSource *interfaces.DataSource, formViewsMap map[string]*FormViewFlag, tables []*interfaces.GetDataTableDetailDataBatchRes,
// // 	dataViewBusinessNameMap map[string]struct{}) {
// // 	defer func() {
// // 		if err := recover(); err != nil {
// // 			logger.Errorf("[ConcurrentCompareView panic] %v", zap.Any("error", err))
// // 		}
// // 	}()

// // 	var codeListIndex int
// // 	for _, table := range tables {
// // 		findView := formViewsMap[table.Name]
// // 		if findView == nil {
// // 			if err := dss.createViews(ctx, dataSource, table, dataViewBusinessNameMap); err != nil {
// // 				return
// // 			}
// // 			codeListIndex++
// // 		} else {
// // 			if err := dss.updateView(ctx, createViewErrorCh, dataSource, findView.DataView, table); err != nil {
// // 				return
// // 			}
// // 			findView.mu.Lock()
// // 			findView.flag = 2
// // 			findView.mu.Unlock()
// // 		}
// // 	}
// // 	if wg != nil {
// // 		wg.Done()
// // 	}
// // }
