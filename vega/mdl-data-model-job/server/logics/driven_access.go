// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import "data-model-job/interfaces"

var (
	EMAccess interfaces.EventModelAccess
	IBAccess interfaces.IndexBaseAccess
	JAccess  interfaces.JobAccess
	KAccess  interfaces.KafkaAccess
	MMAccess interfaces.MetricModelAccess
	UAccess  interfaces.UniqueryAccess
)

func SetJobAccess(jAccess interfaces.JobAccess) {
	JAccess = jAccess
}

func SetKafkaAccess(kAccess interfaces.KafkaAccess) {
	KAccess = kAccess
}

func SetIndexBaseAccess(ibAccess interfaces.IndexBaseAccess) {
	IBAccess = ibAccess
}

func SetUniqueryAccess(uAccess interfaces.UniqueryAccess) {
	UAccess = uAccess
}

func SetMetricModelAccess(mmAccess interfaces.MetricModelAccess) {
	MMAccess = mmAccess
}

func SetEventModelAccess(emAccess interfaces.EventModelAccess) {
	EMAccess = emAccess
}
