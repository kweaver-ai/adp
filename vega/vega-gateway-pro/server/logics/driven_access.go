// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"vega-gateway-pro/interfaces"
)

var (
	DataConnectionAccess interfaces.DataConnectionAccess
	VegaCalculateAccess  interfaces.VegaCalculateAccess
)

func SetDataConnectionAccess(dca interfaces.DataConnectionAccess) {
	DataConnectionAccess = dca
}

func SetVegaViewAccess(vva interfaces.VegaCalculateAccess) {
	VegaCalculateAccess = vva
}
