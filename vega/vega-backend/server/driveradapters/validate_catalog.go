package driveradapters

import (
	"context"

	"vega-backend/interfaces"
)

func ValidateCatalogRequest(ctx context.Context, req *interfaces.CatalogRequest) error {
	if err := validateName(ctx, req.Name); err != nil {
		return err
	}
	if err := ValidateTags(ctx, req.Tags); err != nil {
		return err
	}
	if err := validateDescription(ctx, req.Description); err != nil {
		return err
	}
	return nil
}
