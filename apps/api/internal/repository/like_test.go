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

func TestSetLikeConcurrentIdempotent(t *testing.T) {
	store := newLikeTestStore(t)
	now := time.Now().UTC()
	date := now.Format("2006-01-02")

	const requests = 100
	var waitGroup sync.WaitGroup
	waitGroup.Add(requests)
	errs := make(chan error, requests)
	for index := 0; index < requests; index++ {
		go func() {
			defer waitGroup.Done()
			_, _, err := store.SetLike(1, "/like", "same-visitor", true, date, now)
			errs <- err
		}()
	}
	waitGroup.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("set like: %v", err)
		}
	}

	likes, liked, err := store.GetLikeState(1, "/like", "same-visitor")
	if err != nil {
		t.Fatalf("get like state: %v", err)
	}
	if likes != 1 || !liked {
		t.Fatalf("likes=%d liked=%v, want likes=1 liked=true", likes, liked)
	}
}

func TestSetLikeConcurrentUniqueVisitors(t *testing.T) {
	store := newLikeTestStore(t)
	now := time.Now().UTC()
	date := now.Format("2006-01-02")

	const requests = 100
	var waitGroup sync.WaitGroup
	waitGroup.Add(requests)
	errs := make(chan error, requests)
	for index := 0; index < requests; index++ {
		visitorID := "visitor-" + string(rune('a'+index%26)) + "-" + string(rune('a'+index/26))
		go func() {
			defer waitGroup.Done()
			_, _, err := store.SetLike(1, "/unique", visitorID, true, date, now)
			errs <- err
		}()
	}
	waitGroup.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("set unique like: %v", err)
		}
	}

	likes, _, err := store.GetLikeState(1, "/unique", "visitor-a-a")
	if err != nil {
		t.Fatalf("get like state: %v", err)
	}
	if likes != requests {
		t.Fatalf("likes=%d, want %d", likes, requests)
	}
}

func TestToggleLikeConsistency(t *testing.T) {
	store := newLikeTestStore(t)
	now := time.Now().UTC()
	date := now.Format("2006-01-02")

	const requests = 100
	var waitGroup sync.WaitGroup
	waitGroup.Add(requests)
	errs := make(chan error, requests)
	for index := 0; index < requests; index++ {
		go func() {
			defer waitGroup.Done()
			_, _, err := store.ToggleLike(1, "/toggle", "same-visitor", date, now)
			errs <- err
		}()
	}
	waitGroup.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("toggle like: %v", err)
		}
	}

	likes, liked, err := store.GetLikeState(1, "/toggle", "same-visitor")
	if err != nil {
		t.Fatalf("get toggle state: %v", err)
	}
	if liked && likes != 1 {
		t.Fatalf("liked=true but likes=%d", likes)
	}
	if !liked && likes != 0 {
		t.Fatalf("liked=false but likes=%d", likes)
	}
}

func newLikeTestStore(t *testing.T) *Store {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "likes.db")), &gorm.Config{
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

	if err := db.AutoMigrate(&model.LikeStat{}, &model.LikeRecord{}, &model.LikeDaily{}); err != nil {
		t.Fatalf("migrate like tables: %v", err)
	}
	return New(db)
}
