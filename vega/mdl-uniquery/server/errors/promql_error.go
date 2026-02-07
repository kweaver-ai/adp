// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

import (
	"fmt"
)

type PromQLError struct {
	Typ errorType
	Err error
}

type status string

const (
	StatusSuccess status = "success"
	StatusError   status = "error"
)

type errorType string

const (
	ErrorNone                errorType = ""
	ErrorTimeout             errorType = "timeout"
	ErrorCanceled            errorType = "canceled"
	ErrorExec                errorType = "execution"
	ErrorBadData             errorType = "bad_data"
	ErrorInternal            errorType = "internal"
	ErrorUnavailable         errorType = "unavailable"
	ErrorNotFound            errorType = "not_found"
	ErrorStatusNotAcceptable errorType = "status_not_acceptable"
)

type ResponseError struct {
	Status    status    `json:"status"`
	ErrorType errorType `json:"errorType,omitempty"`
	Error     string    `json:"error,omitempty"`
}

func (err PromQLError) Error() string {
	return fmt.Sprintf("%s", err.Err)
}

func InvalidParamError(err error, parameter string) PromQLError {
	return PromQLError{
		ErrorBadData, fmt.Errorf("invalid parameter %q: %w", parameter, err),
	}
}
