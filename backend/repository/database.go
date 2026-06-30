package repository

import "gorm.io/gorm"

// DB は GORM 接続のラッパー（バックアップ同期など repository 内部用）
type DB struct {
	gorm *gorm.DB
}

func NewDB(gdb *gorm.DB) *DB {
	return &DB{gorm: gdb}
}

func (d *DB) Gorm() *gorm.DB { return d.gorm }
