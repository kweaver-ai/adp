// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

// 数据字典错误码
const (
	// 400
	DataModel_DataDict_BadRequest_ItemsAreEmpty         = "DataModel.DataDict.BadRequest.ItemsAreEmpty"
	DataModel_DataDict_BadRequest_ParseCSVFileFailed    = "DataModel.DataDict.BadRequest.ParseCSVFileFailed"
	DataModel_DataDict_BadRequest_ParseXlsxFileFailed   = "DataModel.DataDict.BadRequest.ParseXlsxFileFailed"
	DataModel_DataDict_DictNameExisted                  = "DataModel.DataDict.DictNameExisted"
	DataModel_DataDict_Duplicated_DictDimension         = "DataModel.DataDict.Duplicated.DictDimension"
	DataModel_DataDict_Duplicated_DictItemKey           = "DataModel.DataDict.Duplicated.DictItemKey"
	DataModel_DataDict_Duplicated_DictItemKeyInFile     = "DataModel.DataDict.Duplicated.DictItemKeyInFile"
	DataModel_DataDict_Duplicated_DictName              = "DataModel.DataDict.Duplicated.DictName"
	DataModel_DataDict_InvalidParameter                 = "DataModel.DataDict.InvalidParameter"
	DataModel_DataDict_InvalidParameter_Dict            = "DataModel.DataDict.InvalidParameter.Dict"
	DataModel_DataDict_InvalidParameter_DictDimension   = "DataModel.DataDict.InvalidParameter.DictDimension"
	DataModel_DataDict_InvalidParameter_DictID          = "DataModel.DataDict.InvalidParameter.DictID"
	DataModel_DataDict_InvalidParameter_DictIDs         = "DataModel.DataDict.InvalidParameter.DictIDs"
	DataModel_DataDict_InvalidParameter_DictItems       = "DataModel.DataDict.InvalidParameter.DictItems"
	DataModel_DataDict_InvalidParameter_DictItemsFile   = "DataModel.DataDict.InvalidParameter.DictItemsFile"
	DataModel_DataDict_InvalidParameter_FileExt         = "DataModel.DataDict.InvalidParameter.FileExt"
	DataModel_DataDict_InvalidParameter_Format          = "DataModel.DataDict.InvalidParameter.Format"
	DataModel_DataDict_InvalidParameter_ImportMode      = "DataModel.DataDict.InvalidParameter.ImportMode"
	DataModel_DataDict_InvalidParameter_ItemID          = "DataModel.DataDict.InvalidParameter.ItemID"
	DataModel_DataDict_InvalidParameter_ItemIDs         = "DataModel.DataDict.InvalidParameter.ItemIDs"
	DataModel_DataDict_LengthExceeded_DictComment       = "DataModel.DataDict.LengthExceeded.DictComment"
	DataModel_DataDict_LengthExceeded_DictDimensionName = "DataModel.DataDict.LengthExceeded.DictDimensionName"
	DataModel_DataDict_LengthExceeded_DictItemComment   = "DataModel.DataDict.LengthExceeded.DictItemComment"
	DataModel_DataDict_LengthExceeded_DictItemKey       = "DataModel.DataDict.LengthExceeded.DictItemKey"
	DataModel_DataDict_LengthExceeded_DictItemValue     = "DataModel.DataDict.LengthExceeded.DictItemValue"
	DataModel_DataDict_LengthExceeded_DictName          = "DataModel.DataDict.LengthExceeded.DictName"
	DataModel_DataDict_NullParameter_DictDimension      = "DataModel.DataDict.NullParameter.DictDimension"
	DataModel_DataDict_NullParameter_DictItemKey        = "DataModel.DataDict.NullParameter.DictItemKey"
	DataModel_DataDict_NullParameter_DictItemValue      = "DataModel.DataDict.NullParameter.DictItemValue"
	DataModel_DataDict_NullParameter_DictName           = "DataModel.DataDict.NullParameter.DictName"

	// 404
	DataModel_DataDict_DictNotFound     = "DataModel.DataDict.DictNotFound"
	DataModel_DataDict_DictItemNotFound = "DataModel.DataDict.DictItemNotFound"

	// 500
	DataModel_DataDict_InternalError                           = "DataModel.DataDict.InternalError"
	DataModel_DataDict_InternalError_BeginTransactionFailed    = "DataModel.DataDict.InternalError.BeginTransactionFailed"
	DataModel_DataDict_InternalError_CommitTransactionFailed   = "DataModel.DataDict.InternalError.CommitTransactionFailed"
	DataModel_DataDict_InternalError_ExportCSVFileFailed       = "DataModel.DataDict.InternalError.ExportCSVFileFailed"
	DataModel_DataDict_InternalError_ExportXlsxFileFailed      = "DataModel.DataDict.InternalError.ExportXlsxFileFailed"
	DataModel_DataDict_InternalError_RollbackTransactionFailed = "DataModel.DataDict.InternalError.RollbackTransactionFailed"
)

var (
	dataDictErrCodeList = []string{

		// ---数据字典模块---
		// 400
		DataModel_DataDict_BadRequest_ItemsAreEmpty,
		DataModel_DataDict_BadRequest_ParseCSVFileFailed,
		DataModel_DataDict_BadRequest_ParseXlsxFileFailed,
		DataModel_DataDict_DictNameExisted,
		DataModel_DataDict_Duplicated_DictDimension,
		DataModel_DataDict_Duplicated_DictItemKey,
		DataModel_DataDict_Duplicated_DictItemKeyInFile,
		DataModel_DataDict_Duplicated_DictName,
		DataModel_DataDict_InvalidParameter,
		DataModel_DataDict_InvalidParameter_Dict,
		DataModel_DataDict_InvalidParameter_DictDimension,
		DataModel_DataDict_InvalidParameter_DictID,
		DataModel_DataDict_InvalidParameter_DictIDs,
		DataModel_DataDict_InvalidParameter_DictItems,
		DataModel_DataDict_InvalidParameter_DictItemsFile,
		DataModel_DataDict_InvalidParameter_FileExt,
		DataModel_DataDict_InvalidParameter_Format,
		DataModel_DataDict_InvalidParameter_ImportMode,
		DataModel_DataDict_InvalidParameter_ItemID,
		DataModel_DataDict_InvalidParameter_ItemIDs,
		DataModel_DataDict_LengthExceeded_DictComment,
		DataModel_DataDict_LengthExceeded_DictDimensionName,
		DataModel_DataDict_LengthExceeded_DictItemComment,
		DataModel_DataDict_LengthExceeded_DictItemKey,
		DataModel_DataDict_LengthExceeded_DictItemValue,
		DataModel_DataDict_LengthExceeded_DictName,
		DataModel_DataDict_NullParameter_DictDimension,
		DataModel_DataDict_NullParameter_DictItemKey,
		DataModel_DataDict_NullParameter_DictItemValue,
		DataModel_DataDict_NullParameter_DictName,

		// 404
		DataModel_DataDict_DictNotFound,
		DataModel_DataDict_DictItemNotFound,

		// 500
		DataModel_DataDict_InternalError,
		DataModel_DataDict_InternalError_BeginTransactionFailed,
		DataModel_DataDict_InternalError_CommitTransactionFailed,
		DataModel_DataDict_InternalError_ExportCSVFileFailed,
		DataModel_DataDict_InternalError_ExportXlsxFileFailed,
		DataModel_DataDict_InternalError_RollbackTransactionFailed,
	}
)
