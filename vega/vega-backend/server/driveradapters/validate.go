package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// 名称合法性校验
func validateName(ctx context.Context, name string) error {
	if name == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaManager_InvalidParameter_Name)
	}

	if utf8.RuneCountInString(name) > interfaces.NAME_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaManager_InvalidParameter_Name).
			WithErrorDetails(fmt.Sprintf("The length of the name %v exceeds %v", name, interfaces.NAME_MAX_LENGTH))
	}

	return nil
}

// tags 的合法性校验
func ValidateTags(ctx context.Context, Tags []string) error {
	if len(Tags) > interfaces.TAGS_MAX_NUMBER {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaManager_InvalidParameter_Tag).
			WithErrorDetails(fmt.Sprintf("The number of tags exceeds %v", interfaces.TAGS_MAX_NUMBER))
	}

	for _, tag := range Tags {
		err := validateTag(ctx, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

// 数据标签名称合法性校验
func validateTag(ctx context.Context, tag string) error {
	// 去除tag的左右空格
	tag = strings.Trim(tag, " ")

	if tag == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaManager_InvalidParameter_Tag)
		// .WithErrorDetails("Data tag name is null")
	}

	if utf8.RuneCountInString(tag) > interfaces.TAG_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaManager_InvalidParameter_Tag).
			WithErrorDetails(fmt.Sprintf("The length of the tag name exceeds %d", interfaces.TAG_MAX_LENGTH))
	}

	if isInvalid := strings.ContainsAny(tag, interfaces.TAG_INVALID_CHARACTER); isInvalid {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaManager_InvalidParameter_Tag).
			WithErrorDetails(fmt.Sprintf("Tag name contains special characters, such as %s", interfaces.TAG_INVALID_CHARACTER))
	}

	return nil
}

// 备注合法性校验
func validateDescription(ctx context.Context, description string) error {
	if utf8.RuneCountInString(description) > interfaces.DESCRIPTION_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaManager_InvalidParameter_Description).
			WithErrorDetails(fmt.Sprintf("The length of the description exceeds %v", interfaces.DESCRIPTION_MAX_LENGTH))
	}
	return nil
}
