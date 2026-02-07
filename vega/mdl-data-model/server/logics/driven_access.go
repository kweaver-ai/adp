// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"database/sql"

	"data-model/interfaces"
)

var (
	DB     *sql.DB
	DCA    interfaces.DataConnectionAccess
	DDA    interfaces.DataDictAccess
	DDIA   interfaces.DataDictItemAccess
	DMJA   interfaces.DataModelJobAccess
	DSA    interfaces.DataSourceAccess
	DVA    interfaces.DataViewAccess
	DVGA   interfaces.DataViewGroupAccess
	DVRCRA interfaces.DataViewRowColumnRuleAccess
	EMA    interfaces.EventModelAccess
	IBA    interfaces.IndexBaseAccess
	MMA    interfaces.MetricModelAccess
	MMGA   interfaces.MetricModelGroupAccess
	MMTA   interfaces.MetricModelTaskAccess
	OMA    interfaces.ObjectiveModelAccess
	PA     interfaces.PermissionAccess
	SRA    interfaces.ScanRecordAccess
	TMA    interfaces.TraceModelAccess
	UA     interfaces.UniqueryAccess
	VGA    interfaces.VegaGatewayAccess
	VMA    interfaces.VegaMetadataAccess
)

func SetDB(db *sql.DB) {
	DB = db
}

func SetDataConnectionAccess(dca interfaces.DataConnectionAccess) {
	DCA = dca
}

func SetDataDictAccess(dda interfaces.DataDictAccess) {
	DDA = dda
}

func SetDataDictItemAccess(ddia interfaces.DataDictItemAccess) {
	DDIA = ddia
}

func SetDataModelJobAccess(dmja interfaces.DataModelJobAccess) {
	DMJA = dmja
}

func SetDataSourceAccess(dsa interfaces.DataSourceAccess) {
	DSA = dsa
}

func SetDataViewAccess(dva interfaces.DataViewAccess) {
	DVA = dva
}

func SetDataViewGroupAccess(dvga interfaces.DataViewGroupAccess) {
	DVGA = dvga
}

func SetDataViewRowColumnRuleAccess(dvrcra interfaces.DataViewRowColumnRuleAccess) {
	DVRCRA = dvrcra
}

func SetEventModelAccess(ema interfaces.EventModelAccess) {
	EMA = ema
}

func SetIndexBaseAccess(iba interfaces.IndexBaseAccess) {
	IBA = iba
}

func SetMetricModelAccess(mma interfaces.MetricModelAccess) {
	MMA = mma
}

func SetMetricModelGroupAccess(mmga interfaces.MetricModelGroupAccess) {
	MMGA = mmga
}

func SetMetricModelTaskAccess(mmta interfaces.MetricModelTaskAccess) {
	MMTA = mmta
}

func SetPermissionAccess(pa interfaces.PermissionAccess) {
	PA = pa
}

func SetScanRecordAccess(sra interfaces.ScanRecordAccess) {
	SRA = sra
}

func SetTraceModelAccess(tma interfaces.TraceModelAccess) {
	TMA = tma
}

func SetUniqueryAccess(ua interfaces.UniqueryAccess) {
	UA = ua
}

func SetObjectiveModelAccess(oma interfaces.ObjectiveModelAccess) {
	OMA = oma
}

func SetVegaGatewayAccess(vga interfaces.VegaGatewayAccess) {
	VGA = vga
}

func SetVegaMetadataAccess(vma interfaces.VegaMetadataAccess) {
	VMA = vma
}
