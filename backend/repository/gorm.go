package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/numduel/numduel/model"
)

func conn(ctx context.Context, gdb *gorm.DB, tx model.Transaction) (*gorm.DB, error) {
	base, err := Conn(gdb.WithContext(ctx), tx)
	if err != nil {
		return nil, err
	}
	return base, nil
}

func forUpdate(gdb *gorm.DB) *gorm.DB {
	return gdb.Clauses(clause.Locking{Strength: "UPDATE"})
}
