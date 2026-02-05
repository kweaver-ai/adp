package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

const (
	ScanMode_Serial     = "serial"
	ScanMode_Concurrent = "concurrent"

	ViewScanStatus_New      = "new"
	ViewScanStatus_Modify   = "modify"
	ViewScanStatus_Delete   = "delete"
	ViewScanStatus_NoChange = "no_change"

	FieldScanStatus_New        = "new"
	FieldScanStatus_Modify     = "modify"
	FieldScanStatus_Delete     = "delete"
	FieldScanStatus_NoChange   = "no_change"
	FieldScanStatus_NotSupport = "not_support"

	FormViewSaveTypeTemp  = "temp"
	FormViewSaveTypeFinal = "final"

	ConcurrentCount        = 10
	GoroutineMinTableCount = 100
)

const (
	DataSourceAvailable = "avaliable"
	DataSourceScanning  = "scanning"

	DataSourceType_Excel      = "excel"
	DataSourceType_TingYun    = "tingyun"
	DataSourceType_AS7        = "anyshare7"
	DataSourceType_IndexBase  = "index_base"
	DataSourceType_OpenSearch = "opensearch"

	DataSourceID_IndexBase = "cedb5294-07c3-45b1-a273-17baefa62800"
)

const (
	DefaultSchema     = "default"
	ManagementScanner = "management_scanner"
)

const (
	// 字段的advanced params
	FieldAdvancedParams_VirtualDataType = "virtualFieldType"
	FieldAdvancedParams_OriginFieldType = "originFieldType"
	FieldAdvancedParams_IsNullable      = "IS_NULLABLE"
	FieldAdvancedParams_ColumnDef       = "COLUMN_DEF"
	FieldAdvancedParams_CheckPrimaryKey = "checkPrimaryKey"
	FieldAdvancedParams_MappingConfig   = "mappingConfig"

	// 表的advanced params
	TableAdvancedParams_ExcelSheet            = "sheet"
	TableAdvancedParams_ExcelStartCell        = "startCell"
	TableAdvancedParams_ExcelEndCell          = "endCell"
	TableAdvancedParams_ExcelHasHeaders       = "hasHeaders"
	TableAdvancedParams_ExcelSheetAsNewColumn = "sheetAsNewColumn"
	TableAdvancedParams_ExcelFileName         = "fileName"
)

const UUIDRegexString = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"

const (
	FormViewPublicTopic = "af.data-view.es-index"
	EntityChangeTopic   = "af.business-grooming.entity_change"
)
const UnallocatedId = "00000000-0000-0000-0000-000000000000"

var CustomViewSource = "custom_view_source"
var LogicEntityViewSource = "logic_entity_view_source"

var ViewNeedRecreate = "is stale; it must be re-created`"

var SampleDataCount = "sample_data_count"
var SampleDataType = "sample_data_type"
var Synthetic = "synthetic"
var Real = "real"

var CodeGenerationRuleUUIDDataView = uuid.MustParse("13daf448-d9c4-11ee-81aa-005056b4b3fc")

type View struct {
	Id           string `json:"id"`
	BusinessName string `json:"business_name"`
}

type ErrorView struct {
	Id            string          `json:"id"`
	TechnicalName string          `json:"technical_name"`
	Error         *rest.HTTPError `json:"error"`
}
type ScanTask struct {
	DataSourceID string `json:"data_source_id" binding:"required"`
	TaskID       string `json:"task_id"  form:"task_id" binding:"omitempty"`
	ProjectID    string `json:"project_id"  form:"project_id" binding:"omitempty"`
}

type ScanResult struct {
	ErrorView      []*ErrorView `json:"error_views"`
	ErrorViewCount int          `json:"error_view_count"`
	ScanViewCount  int          `json:"scan_view_count"`
}

type FinishProjectReq struct {
	FinishProjectReqParamPath `param_type:"body"`
}

type FinishProjectReqParamPath struct {
	TaskIDs []string `json:"task_id" form:"task_id" binding:"required,dive,uuid" example:"88f78432-ee4e-43df-804c-4ccc4ff17f15"`
}

type ListDataSourceQueryParams struct {
	QueryType string
}

//go:generate mockgen -source ../interfaces/data_source_service.go -destination ../interfaces/mock/mock_data_source_service.go
type DataSourceService interface {
	Scan(ctx context.Context, req *ScanTask) (*ScanResult, error)
	ListDataSourcesWithScanRecord(ctx context.Context, queryParams *ListDataSourceQueryParams) (*ListDataSourcesResult, error)
}
