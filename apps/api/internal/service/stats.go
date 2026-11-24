package service

import (
	"fmt"
)

type PublicPageStats struct {
	Path        string `json:"path"`
	Comments    int64  `json:"comments"`
	Views       uint64 `json:"views"`
	UniqueViews uint64 `json:"uniqueViews"`
	Likes       uint64 `json:"likes"`
	Liked       bool   `json:"liked"`
}

func (s *Service) GetPublicPageStats(siteID uint, rawPaths []string, visitorID string) ([]PublicPageStats, error) {
	paths, err := normalizePathList(rawPaths, s.view.MaxBatchPaths)
	if err != nil {
		return nil, err
	}
	visitorID, err = normalizeVisitorID(visitorID)
	if err != nil {
		return nil, err
	}

	viewStats, err := s.repository.GetPageStats(siteID, paths)
	if err != nil {
		return nil, fmt.Errorf("get page stats: %w", err)
	}

	items := make([]PublicPageStats, 0, len(paths))
	for _, path := range paths {
		comments, err := s.CountApprovedComments(siteID, path)
		if err != nil {
			return nil, fmt.Errorf("count comments: %w", err)
		}
		likes, liked, err := s.repository.GetLikeState(siteID, path, visitorID)
		if err != nil {
			return nil, fmt.Errorf("get like state: %w", err)
		}

		item := PublicPageStats{
			Path:     path,
			Comments: comments,
			Likes:    likes,
			Liked:    liked,
		}
		if stat, ok := viewStats[path]; ok {
			item.Views = stat.Views
			item.UniqueViews = stat.UniqueViews
		}
		items = append(items, item)
	}
	return items, nil
}
