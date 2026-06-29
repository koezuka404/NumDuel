package repository

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func forUpdate(gdb *gorm.DB) *gorm.DB {
	return gdb.Clauses(clause.Locking{Strength: "UPDATE"})
}
