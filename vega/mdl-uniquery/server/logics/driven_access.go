// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"uniquery/interfaces"
)

var (
	DCAccess    interfaces.DataConnectionAccess
	DDAccess    interfaces.DataDictAccess
	DDService   interfaces.DataDictService
	DMAccess    interfaces.DataManagerAccess
	DVAccess    interfaces.DataViewAccess
	DVRCRAccess interfaces.DataViewRowColumnRuleAccess
	EMAccess    interfaces.EventModelAccess
	IBAccess    interfaces.IndexBaseAccess
	KAccess     interfaces.KafkaAccess
	LGAccess    interfaces.LogGroupAccess
	MMAccess    interfaces.MetricModelAccess
	OMAccess    interfaces.ObjectiveModelAccess
	OSAccess    interfaces.OpenSearchAccess
	PA          interfaces.PermissionAccess
	SAccess     interfaces.SearchAccess
	StAccess    interfaces.StaticAccess
	TMAccess    interfaces.TraceModelAccess
	VDSAccess   interfaces.VegaDataSourceAccess
	VGAccess    interfaces.VegaGatewayAccess
	VVA         interfaces.VegaAccess
)

func SetDataConnectionAccess(dca interfaces.DataConnectionAccess) {
	DCAccess = dca
}

func SetDataDictAccess(dda interfaces.DataDictAccess) {
	DDAccess = dda
}

func SetDataDictService(dds interfaces.DataDictService) {
	DDService = dds
}

func SetDataManagerAccess(dma interfaces.DataManagerAccess) {
	DMAccess = dma
}
func SetDataViewAccess(dva interfaces.DataViewAccess) {
	DVAccess = dva
}

func SetDataViewRowColumnRuleAccess(dvrcra interfaces.DataViewRowColumnRuleAccess) {
	DVRCRAccess = dvrcra
}

func SetEventModelAccess(o interfaces.EventModelAccess) {
	EMAccess = o
}

func SetIndexBaseAccess(iba interfaces.IndexBaseAccess) {
	IBAccess = iba
}

func SetKafkaAccess(kAccess interfaces.KafkaAccess) {
	KAccess = kAccess
}

func SetLogGroupAccess(lga interfaces.LogGroupAccess) {
	LGAccess = lga
}

func SetMetricModelAccess(mma interfaces.MetricModelAccess) {
	MMAccess = mma
}

func SetObjectiveModelAccess(o interfaces.ObjectiveModelAccess) {
	OMAccess = o
}

func SetOpenSearchAccess(osa interfaces.OpenSearchAccess) {
	OSAccess = osa
}

func SetPermissionAccess(pa interfaces.PermissionAccess) {
	PA = pa
}

func SetSearchAccess(sa interfaces.SearchAccess) {
	SAccess = sa
}

func SetStaticAccess(sa interfaces.StaticAccess) {
	StAccess = sa
}

func SetTraceModelAccess(tma interfaces.TraceModelAccess) {
	TMAccess = tma
}

func SetVegaDataSourceAccess(vdsa interfaces.VegaDataSourceAccess) {
	VDSAccess = vdsa
}

func SetVegaGatewayAccess(vga interfaces.VegaGatewayAccess) {
	VGAccess = vga
}

func SetVegaViewAccess(vva interfaces.VegaAccess) {
	VVA = vva
}
