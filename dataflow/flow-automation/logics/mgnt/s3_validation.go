package mgnt

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/log"
)

// ValidateS3Config 验证S3配置
func (m *mgnt) ValidateS3Config(ctx context.Context, bucket, path string) (*S3ValidationResult, error) {
	result := &S3ValidationResult{}

	// 检查S3适配器是否已初始化
	if m.s3Adapter == nil {
		return nil, fmt.Errorf("S3 adapter not initialized, please configure S3 credentials")
	}

	// 验证Bucket
	if err := m.s3Adapter.ValidateBucket(ctx, bucket); err != nil {
		result.BucketExists = false
		result.Message = fmt.Sprintf("Bucket validation failed: %v", err)
		log.Warnf("[mgnt.ValidateS3Config] Bucket validation failed: %v", err)
		return result, nil
	}
	result.BucketExists = true

	// 验证路径并列出对象
	objects, err := m.s3Adapter.ListObjects(ctx, bucket, path)
	if err != nil {
		result.PathAccessible = false
		result.Message = fmt.Sprintf("Path access failed: %v", err)
		log.Warnf("[mgnt.ValidateS3Config] Path access failed: %v", err)
		return result, nil
	}
	result.PathAccessible = true
	result.FileCount = len(objects)
	result.Message = fmt.Sprintf("Successfully validated. Found %d objects", len(objects))

	log.Infof("[mgnt.ValidateS3Config] Validation successful for bucket=%s, path=%s, count=%d",
		bucket, path, len(objects))
	return result, nil
}
