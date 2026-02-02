package driveradapters

import (
	"context"
	"vega-manager/interfaces"
)

func ValidateResourceRequest(ctx context.Context, req *interfaces.ResourceRequest) error {
	if err := validateName(ctx, req.Name); err != nil {
		return err
	}
	if err := ValidateTags(ctx, req.Tags); err != nil {
		return err
	}
	if err := validateComment(ctx, req.Comment); err != nil {
		return err
	}
	return nil
}
