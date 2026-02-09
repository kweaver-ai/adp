// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"database/sql"
	"sync"

	sq "github.com/Masterminds/squirrel"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	_ "github.com/kweaver-ai/proton-rds-sdk-go/driver"

	"uniquery/common"
	"uniquery/interfaces"
)

const (
	STATIC_METRIC_INDEX = "t_static_metric_index"
)

var (
	stAccessOnce sync.Once
	stAccess     interfaces.StaticAccess
)

type staticAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewStaticAccess(appSetting *common.AppSetting) interfaces.StaticAccess {
	stAccessOnce.Do(func() {
		stAccess = &staticAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return stAccess
}

// 获取指标索引库的时间静态分割表数据
// 批量获取指标模型信息，不包括任务信息。任务信息单独查询
func (sa *staticAccess) GetIndexBaseSplitTime() ([]interfaces.IndexBaseSplitTime, error) {
	indexBaseSplitTimes := make([]interfaces.IndexBaseSplitTime, 0)
	//查询
	sqlStr, vals, err := sq.Select(
		"f_base_type",
		"f_split_time").
		From(STATIC_METRIC_INDEX).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select model by id, error: %s", err.Error())
		return indexBaseSplitTimes, err
	}

	// 记录处理的 sql 字符串
	rows, err := sa.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		return indexBaseSplitTimes, err
	}
	defer rows.Close()
	for rows.Next() {
		indexBaseSplitTime := interfaces.IndexBaseSplitTime{}

		err := rows.Scan(
			&indexBaseSplitTime.BaseType,
			&indexBaseSplitTime.SplitTime,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			return indexBaseSplitTimes, err
		}

		indexBaseSplitTimes = append(indexBaseSplitTimes, indexBaseSplitTime)
	}

	return indexBaseSplitTimes, nil
}
