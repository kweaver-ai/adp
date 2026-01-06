package interfaces

import (
	"context"
	"time"
)

//go:generate mockgen -source ../interfaces/fetch_service.go -destination ../interfaces/mock/mock_fetch_service.go
type FetchService interface {
	FetchQuery(ctx context.Context, query *FetchQueryReq) (*FetchResp, error)
	NextQuery(ctx context.Context, query *NextQueryReq) (*FetchResp, error)
}

type FetchQueryReq struct {
	DataSourceId string `json:"data_source_id" binding:"required" example:""`                 // 数据源ID
	Type         int    `json:"type" binding:"required,oneof=1 2" example:"1"`                // 查询类型 1-同步查询 2-流式查询
	Sql          string `json:"sql" binding:"required" example:"select * from table"`         // 查询语句
	BatchSize    *int   `json:"batch_size" binding:"omitempty,min=1,max=10000" example:"100"` // 批次大小
	Timeout      *int   `json:"timeout" binding:"omitempty,min=1,max=1800" example:"60"`      // 超时时间（秒）
}

type NextQueryReq struct {
	QueryId   string `uri:"query_id" binding:"required" example:"123456"`                  // 查询ID
	Slug      string `uri:"slug" binding:"required" example:"x123456"`                     //
	Token     int    `uri:"token" binding:"required" example:"1"`                          // 分页令牌
	BatchSize int    `form:"batch_size" binding:"omitempty,min=1,max=10000" example:"100"` // 批次大小
}

type FetchResp struct {
	NextUri    string    `json:"next_uri,omitempty"`
	Columns    []*Column `json:"columns"`
	Entries    []*[]any  `json:"entries"`
	TotalCount int64     `json:"total_count"`
}

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type QueryResult struct {
	ResultSet any       // 存储结果集，具体类型根据连接器实现而定
	Columns   []*Column `json:"columns"`
	Data      []*[]any  `json:"data"`
}

type ResultCache struct {
	ResultSet     any         // 存储结果集，具体类型根据连接器实现而定
	Token         int         // 查询下标
	Columns       []*Column   // 列信息
	ResultChan    chan *[]any // 查询结果通道
	Error         error       // 存储查询错误
	MaxExceedTime time.Time   // 查询最大超时时间
}
