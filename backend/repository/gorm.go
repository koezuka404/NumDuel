package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/numduel/numduel/model"
)

type baseRepo struct {
	db *gorm.DB
}

func mapDBError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func mapRowsAffected(rows int64, err error) error {
	if err != nil {
		return err
	}
	if rows == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

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
