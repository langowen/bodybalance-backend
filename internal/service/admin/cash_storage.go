package admin

import (
	"context"
)

type CashStorage interface {
	InvalidateVideosCache(ctx context.Context) error
	InvalidateCategoriesCache(ctx context.Context) error
	InvalidateAccountsCache(ctx context.Context) error
	InvalidateAllCache(ctx context.Context) error
}
