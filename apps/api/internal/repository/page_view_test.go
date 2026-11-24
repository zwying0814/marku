package repository

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"marku/api/internal/model"
)

func TestRecordPageViewConcurrency(t *testing.T) {
	store := newViewTestStore(t)
	now := time.Now().UTC()
	date := now.Format("2006-01-02")

	const requests = 100
	var waitGroup sync.WaitGroup
	waitGroup.Add(requests)
	errs := make(chan error, requests)

	for index := 0; index < requests; index++ {
		go func() {
			defer waitGroup.Done()
			_, _, err := store.RecordPageView(1, "/concurrent", date, "", false, now)
			errs <- err
		}()
	}
	waitGroup.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent record: %v", err)
		}
	}

	stat, err := store.FindPageStat(1, "/concurrent")
	if err != nil {
		t.Fatalf("find page stat: %v", err)
	}
	if stat.Views != requests {
		t.Fatalf("views = %d, want %d", stat.Views, requests)
	}
}

func TestRecordPageViewUniqueDedupe(t *testing.T) {
	store := newViewTestStore(t)
	now := time.Now().UTC()
	date := now.Format("2006-01-02")

	const requests = 20
	var waitGroup sync.WaitGroup
	waitGroup.Add(requests)
	errs := make(chan error, requests)

	for index := 0; index < requests; index++ {
		go func() {
			defer waitGroup.Done()
			_, _, err := store.RecordPageView(1, "/unique", date, "same-visitor", true, now)
			errs <- err
		}()
	}
	waitGroup.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("dedupe record: %v", err)
		}
	}

	stat, err := store.FindPageStat(1, "/unique")
	if err != nil {
		t.Fatalf("find page stat: %v", err)
	}
	if stat.Views != requests || stat.UniqueViews != 1 {
		t.Fatalf("views=%d unique=%d, want views=%d unique=1", stat.Views, stat.UniqueViews, requests)
	}
}

func newViewTestStore(t *testing.T) *Store {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "views.db")), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql database: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(&model.PageStat{}, &model.PageViewDaily{}, &model.PageViewVisitor{}); err != nil {
		t.Fatalf("migrate view tables: %v", err)
	}
	return New(db)
}
