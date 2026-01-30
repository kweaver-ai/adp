package mgnt

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/log"
)

const (
	// S3ListObjectsOperator S3列表对象操作符
	S3ListObjectsOperator = "@s3/list-objects"
)

// S3DataItem S3数据项
type S3DataItem struct {
	Bucket       string `json:"bucket"`
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
}

// handleS3DataSource 处理S3数据源
func (m *mgnt) handleS3DataSource(ctx context.Context, dag *entity.Dag, dagIns *entity.DagInstance) error {
	if dag.TriggerConfig == nil || dag.TriggerConfig.DataSource == nil {
		return fmt.Errorf("trigger config or data source is nil")
	}

	dataSource := dag.TriggerConfig.DataSource
	if dataSource.Operator != S3ListObjectsOperator {
		return fmt.Errorf("unsupported data source operator: %s", dataSource.Operator)
	}

	// 解析S3配置参数
	if dataSource.Parameters == nil {
		return fmt.Errorf("data source parameters is nil")
	}

	sources := dataSource.Parameters.Sources
	if len(sources) == 0 {
		return fmt.Errorf("no S3 sources configured")
	}

	// 获取S3适配器
	s3Adapter := m.s3Adapter
	if s3Adapter == nil {
		return fmt.Errorf("S3 adapter not initialized")
	}

	// 收集所有S3对象
	var allItems []S3DataItem
	for _, source := range sources {
		// 验证Bucket
		if err := s3Adapter.ValidateBucket(ctx, source.Bucket); err != nil {
			log.Warnf("[mgnt.handleS3DataSource] Failed to validate bucket %s: %v", source.Bucket, err)
			continue
		}

		// 列出对象
		objects, err := s3Adapter.ListObjects(ctx, source.Bucket, source.Path)
		if err != nil {
			log.Warnf("[mgnt.handleS3DataSource] Failed to list objects in bucket %s with prefix %s: %v",
				source.Bucket, source.Path, err)
			continue
		}

		// 转换为数据项
		for _, obj := range objects {
			allItems = append(allItems, S3DataItem{
				Bucket:       source.Bucket,
				Key:          obj.Key,
				Size:         obj.Size,
				LastModified: obj.LastModified.Format("2006-01-02T15:04:05Z"),
			})
		}
	}

	if len(allItems) == 0 {
		return fmt.Errorf("no objects found in S3 sources")
	}

	// 将S3对象列表存储到共享数据中
	// 触发器步骤可以访问 __0 来获取数据项列表
	dagIns.ShareData.Set("0", map[string]interface{}{
		"items": allItems,
		"count": len(allItems),
	})

	log.Infof("[mgnt.handleS3DataSource] Successfully loaded %d objects from S3", len(allItems))
	return nil
}
