package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"gorm.io/gorm"

	"marku/api/internal/config"
	"marku/api/internal/database"
	"marku/api/internal/logging"
	"marku/api/internal/model"
	"marku/api/internal/repository"
	"marku/api/internal/security"
	"marku/api/internal/service"
)

type demoSite struct {
	Name  string
	Slug  string
	Paths []string
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
	slog.SetDefault(logging.New(cfg.Log.Level, cfg.Log.Format))

	db, err := database.Open(cfg)
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(db); err != nil {
			slog.Error("close database", "error", err)
		}
	}()

	if err := database.AutoMigrate(db, model.MigrationModels()...); err != nil {
		slog.Error("run migrations", "error", err)
		os.Exit(1)
	}

	store := repository.New(db)
	services := service.New(store, cfg)
	if created, err := services.SeedAdmin(); err != nil {
		slog.Error("seed initial admin", "error", err)
		os.Exit(1)
	} else if created {
		slog.Info("initial admin created", "email", cfg.Auth.AdminEmail)
	}

	demos := []demoSite{
		{Name: "Demo Blog", Slug: "demo-blog", Paths: []string{"/posts/hello", "/posts/second"}},
		{Name: "Demo Docs", Slug: "demo-docs", Paths: []string{"/docs/getting-started"}},
	}

	for _, demo := range demos {
		site, key, err := ensureSite(store, services, demo)
		if err != nil {
			slog.Error("seed site", "slug", demo.Slug, "error", err)
			os.Exit(1)
		}

		if key != "" {
			slog.Info("demo site created", "name", site.Name, "slug", site.Slug, "site_key", key)
		} else {
			slog.Info("demo site already exists", "name", site.Name, "slug", site.Slug, "site_key_prefix", site.SiteKeyPrefix)
		}

		seedComments(store, site, demo.Paths)
		seedViews(store, site, demo.Paths)
		seedLikes(store, site, demo.Paths)
	}

	slog.Info("demo data seeded")
}

func ensureSite(store *repository.Store, services *service.Service, demo demoSite) (*model.Site, string, error) {
	site, err := store.FindSiteBySlug(demo.Slug)
	if err == nil {
		return site, "", nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", fmt.Errorf("find site %q: %w", demo.Slug, err)
	}

	site, key, err := services.CreateSite(service.CreateSiteInput{
		Name:              demo.Name,
		Slug:              demo.Slug,
		CommentModeration: model.CommentModerationApproved,
	})
	if err != nil {
		return nil, "", fmt.Errorf("create site %q: %w", demo.Slug, err)
	}
	return site, key, nil
}

func seedComments(store *repository.Store, site *model.Site, paths []string) {
	for _, path := range paths {
		existing, err := store.ListApprovedCommentsByPath(site.ID, path)
		if err != nil {
			slog.Warn("check demo comments", "path", path, "error", err)
			continue
		}
		if len(existing) > 0 {
			continue
		}

		root := &model.Comment{
			SiteID:      site.ID,
			Path:        path,
			AuthorName:  "Alice",
			AuthorEmail: "alice@example.com",
			Content:     "This demo comment is seeded by Marku.",
			ContentHash: security.HashContent("This demo comment is seeded by Marku."),
			Status:      model.CommentStatusApproved,
			UserAgent:   "seed",
		}
		if err := store.CreateComment(root); err != nil {
			slog.Warn("create demo comment", "path", path, "error", err)
			continue
		}

		reply := &model.Comment{
			SiteID:      site.ID,
			Path:        path,
			ParentID:    &root.ID,
			AuthorName:  "Admin",
			AuthorEmail: "admin@example.com",
			Content:     "Thanks for reading the demo.",
			ContentHash: security.HashContent("Thanks for reading the demo."),
			Status:      model.CommentStatusApproved,
			IsAdmin:     true,
			UserAgent:   "seed",
		}
		if err := store.CreateComment(reply); err != nil {
			slog.Warn("create demo reply", "path", path, "error", err)
		}
	}
}

func seedViews(store *repository.Store, site *model.Site, paths []string) {
	now := time.Now().UTC()
	date := now.Format("2006-01-02")

	for _, path := range paths {
		if _, err := store.FindPageStat(site.ID, path); err == nil {
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Warn("check demo views", "path", path, "error", err)
			continue
		}

		for index := 0; index < 5; index++ {
			if _, _, err := store.RecordPageView(site.ID, path, date, "", false, now); err != nil {
				slog.Warn("seed demo views", "path", path, "error", err)
				break
			}
		}
	}
}

func seedLikes(store *repository.Store, site *model.Site, paths []string) {
	now := time.Now().UTC()
	date := now.Format("2006-01-02")

	for index, path := range paths {
		for visitor := 0; visitor < 3; visitor++ {
			visitorID := fmt.Sprintf("seed-visitor-%d-%d", index, visitor)
			if _, _, err := store.SetLike(site.ID, path, visitorID, true, date, now); err != nil {
				slog.Warn("seed demo likes", "path", path, "error", err)
			}
		}
	}
}
