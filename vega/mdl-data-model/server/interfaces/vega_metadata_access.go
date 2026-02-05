package interfaces

import (
	"context"
	"fmt"
)

type ListMetadataTablesParams struct {
	DataSourceId string
	Keyword      string // 模糊搜索表名称
	UpdateTime   string // 更新时间
	PaginationQueryParameters
}

type SimpleMetadataTable struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	UpdateTime string `json:"update_time"`
}

type MetadataTable struct {
	TableID    string         `json:"table_id"`
	Table      *TableInfo     `json:"table"`
	DataSource DataSourceInfo `json:"datasource"`
	FieldList  []*MetaField   `json:"field_list"`
}

type TableInfo struct {
	ID             string              `json:"table_id"`
	Name           string              `json:"table_name"`
	AdvancedParams AdvancedParamStruct `json:"table_advanced_params"`
	Description    string              `json:"table_description"`
	Rows           int64               `json:"table_rows"`
}

type DataSourceInfo struct {
	DataSourceID   string `json:"ds_id"`
	DataSourceName string `json:"ds_name"`
	Type           string `json:"ds_type"`
	Catalog        string `json:"ds_catalog"`
	Database       string `json:"ds_database"`
	Schema         string `json:"ds_schema"`
}

type MetaField struct {
	FieldName      string              `json:"f_field_name"`
	TableID        string              `json:"f_table_id"`
	TableName      string              `json:"f_table_name"`
	FieldType      string              `json:"f_field_type"`
	FieldLength    int32               `json:"f_field_length"`
	FieldPrecision int32               `json:"f_field_precision"`
	FieldComment   string              `json:"f_field_comment"`
	AdvancedParams AdvancedParamStruct `json:"f_advanced_params"`
}

func (fb *MetaField) String() string {
	return fmt.Sprintf("MetaField{name: %s, type: %s, comment: %s}", fb.FieldName, fb.FieldType, fb.FieldComment)
}

type AdvancedParamStruct []*AdvancedParams

//go:generate mockgen -source ../interfaces/vega_metadata_access.go -destination ../interfaces/mock/mock_vega_metadata_access.go
type VegaMetadataAccess interface {
	ListMetadataTablesBySourceID(ctx context.Context, params *ListMetadataTablesParams) ([]SimpleMetadataTable, error)
	GetMetadataTablesByIDs(ctx context.Context, tableIDs []string) ([]MetadataTable, error)
}

// 判断key是否存在
func (s AdvancedParamStruct) HasKey(key string) bool {
	for _, params := range s {
		if params.Key == key {
			return true
		}
	}
	return false
}

// GetValue 获取key对应的value，若value为空，则根据key返回默认值
func (s AdvancedParamStruct) GetValue(key string) any {
	for _, params := range s {
		if params.Key == key {
			val := params.Value
			if val == nil {
				return s.getDefaultValue(key)
			}

			return val
		}
	}

	// 如果没有找到key，也根据key返回各自的默认值
	return s.getDefaultValue(key)
}

func (s AdvancedParamStruct) getDefaultValue(key string) any {
	switch key {
	case FieldAdvancedParams_VirtualDataType:
		return ""
	case FieldAdvancedParams_OriginFieldType:
		return ""
	case FieldAdvancedParams_IsNullable:
		return ""
	case FieldAdvancedParams_ColumnDef:
		return ""
	case FieldAdvancedParams_CheckPrimaryKey:
		return ""
	case FieldAdvancedParams_MappingConfig:
		return map[string]any{}
	case TableAdvancedParams_ExcelSheet:
		return ""
	case TableAdvancedParams_ExcelStartCell:
		return ""
	case TableAdvancedParams_ExcelEndCell:
		return ""
	case TableAdvancedParams_ExcelHasHeaders:
		return false
	case TableAdvancedParams_ExcelSheetAsNewColumn:
		return false
	case TableAdvancedParams_ExcelFileName:
		return ""
	default:
		return ""
	}
}

func (s AdvancedParamStruct) IsPrimaryKey() bool {
	for _, params := range s {
		if params.Key == FieldAdvancedParams_CheckPrimaryKey && params.Value == "YES" {
			return true
		}
		if params.Key == FieldAdvancedParams_CheckPrimaryKey && params.Value == "NO" {
			return false
		}
	}
	return false
}
