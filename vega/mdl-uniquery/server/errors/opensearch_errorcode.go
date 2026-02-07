// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

import (
	"net/http"
)

const (
	ModuleName = "UniQuery"

	RequestParseException    = ModuleName + ".RequestParseException"
	UnauthorizedException    = ModuleName + ".UnauthorizedException"
	InternalServerError      = ModuleName + ".InternalServerError"
	IllegalArgumentException = ModuleName + ".IllegalArgumentException"
)

var (
	openSearchErrors = map[string]OpenSearchError{
		RequestParseException: {
			StatusCode: http.StatusBadRequest,
			ErrBody: &errBody{
				ErrType: RequestParseException,
			},
		},
		UnauthorizedException: {
			StatusCode: http.StatusUnauthorized,
			ErrBody: &errBody{
				ErrType: UnauthorizedException,
			},
		},
		InternalServerError: {
			StatusCode: http.StatusInternalServerError,
			ErrBody: &errBody{
				ErrType: InternalServerError,
				Reason:  "internal server error",
			},
		},
		IllegalArgumentException: {
			StatusCode: http.StatusBadRequest,
			ErrBody: &errBody{
				ErrType: IllegalArgumentException,
				Reason:  "illegal_argument_exception",
			},
		},
	}
)
