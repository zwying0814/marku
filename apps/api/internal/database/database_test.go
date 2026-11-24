package database

import (
	"path/filepath"
	"testing"
	"time"

	"marku/api/internal/config"
)

type migrationTestModel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:64"`
}

func TestOpenAndMigrateSQLite(t *testing.T) {
	cfg := config.Config{
		DB: config.DBConfig{
			Driver:          "sqlite",
			DSN:             filepath.Join(t.TempDir(), "marku-test.db"),
			MaxOpenConns:    1,
			MaxIdleConns:    1,
			ConnMaxLifetime: time.Minute,
		},
	}

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() {
		if err := Close(db); err != nil {
			t.Fatalf("close sqlite: %v", err)
		}
	}()

	if err := Ping(db); err != nil {
		t.Fatalf("ping sqlite: %v", err)
	}
	if err := AutoMigrate(db, &migrationTestModel{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if !db.Migrator().HasTable(&migrationTestModel{}) {
		t.Fatal("expected migration test table to exist")
	}
}

func TestOpenRejectsUnsupportedDriver(t *testing.T) {
	cfg := config.Config{
		DB: config.DBConfig{
			Driver:          "oracle",
			DSN:             "test",
			ConnMaxLifetime: time.Minute,
		},
	}
	if _, err := Open(cfg); err == nil {
		t.Fatal("expected unsupported driver error")
	}
}
