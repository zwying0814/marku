package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"marku/api/internal/config"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	dialector, err := dialectorFor(cfg.DB)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormlogger.New(
			log.New(os.Stdout, "", log.LstdFlags),
			gormlogger.Config{
				SlowThreshold:             500 * time.Millisecond,
				LogLevel:                  gormlogger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open %s database: %w", cfg.DB.Driver, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql database handle: %w", err)
	}

	if cfg.DB.Driver == "sqlite" {
		// SQLite serializes writes; one connection keeps local development predictable.
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	} else {
		sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	}
	sqlDB.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping %s database: %w", cfg.DB.Driver, err)
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB, models ...any) error {
	if len(models) == 0 {
		return nil
	}
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	return nil
}

func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql database handle: %w", err)
	}
	return sqlDB.Ping()
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql database handle: %w", err)
	}
	return sqlDB.Close()
}

func dialectorFor(cfg config.DBConfig) (gorm.Dialector, error) {
	switch cfg.Driver {
	case "sqlite":
		if err := ensureSQLiteDir(cfg.DSN); err != nil {
			return nil, err
		}
		return sqlite.Open(cfg.DSN), nil
	case "mysql":
		return mysql.Open(cfg.DSN), nil
	case "postgres":
		return postgres.Open(cfg.DSN), nil
	default:
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
	}
}

func ensureSQLiteDir(dsn string) error {
	path := strings.TrimPrefix(dsn, "file:")
	if index := strings.Index(path, "?"); index >= 0 {
		path = path[:index]
	}
	if path == "" || path == ":memory:" {
		return nil
	}

	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create database directory %q: %w", dir, err)
	}
	return nil
}
