package entity

import "context"

type ZipCodeRepo interface {
	ValidateZipCode(ctx context.Context, zipCode string) bool
}
