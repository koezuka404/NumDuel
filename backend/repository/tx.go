package repository

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

func WithTx(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

func dbFromCtx(ctx context.Context, db *gorm.DB) *gorm.DB {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	if ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}
