package repository

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNotFound = errors.New("not found")

func mapRecordNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func firstOne[T any](db *gorm.DB, dest *T) error {
	return mapRecordNotFound(db.First(dest).Error)
}

func firstOneForUpdate[T any](db *gorm.DB, dest *T) error {
	return mapRecordNotFound(
		db.Clauses(clause.Locking{Strength: "UPDATE"}).First(dest).Error,
	)
}

func findOptional[T any](db *gorm.DB) (*T, error) {
	var row T
	err := db.First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func findOptionalForUpdate[T any](db *gorm.DB) (*T, error) {
	var row T
	err := firstOneForUpdate(db, &row)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func paginatePage(page, limit int) (int, int) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	return limit, (page - 1) * limit
}

func rowsAffected(err error, n int64) error {
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
