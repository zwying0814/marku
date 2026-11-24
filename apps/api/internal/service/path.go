package service

import (
	"fmt"
	"strings"
)

func normalizePagePath(raw string) (string, error) {
	path := strings.TrimSpace(raw)
	if path == "" {
		return "", fmt.Errorf("%w: path is required", ErrInvalidPagePath)
	}
	if index := strings.IndexAny(path, "?#"); index >= 0 {
		path = path[:index]
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}
	if len(path) > 512 {
		return "", fmt.Errorf("%w: path is too long", ErrInvalidPagePath)
	}
	if strings.ContainsAny(path, "\r\n\t") {
		return "", fmt.Errorf("%w: path contains invalid characters", ErrInvalidPagePath)
	}
	if path == "" {
		path = "/"
	}
	return path, nil
}

func normalizePathList(rawPaths []string, max int) ([]string, error) {
	if len(rawPaths) == 0 {
		return nil, fmt.Errorf("%w: paths is required", ErrInvalidViewInput)
	}
	if len(rawPaths) > max {
		return nil, fmt.Errorf("%w: at most %d paths are allowed", ErrInvalidViewInput, max)
	}

	paths := make([]string, 0, len(rawPaths))
	seen := make(map[string]struct{}, len(rawPaths))
	for _, rawPath := range rawPaths {
		if strings.TrimSpace(rawPath) == "" {
			continue
		}
		path, err := normalizePagePath(rawPath)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("%w: paths is required", ErrInvalidViewInput)
	}
	return paths, nil
}
