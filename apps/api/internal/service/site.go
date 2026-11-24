package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"

	"marku/api/internal/model"
	"marku/api/internal/security"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type CreateSiteInput struct {
	Name              string
	Slug              string
	Domain            string
	AllowedOrigins    []string
	CommentModeration string
}

type UpdateSiteInput struct {
	Name              *string
	Domain            *string
	Status            *string
	AllowedOrigins    *[]string
	CommentModeration *string
}

func (s *Service) CreateSite(input CreateSiteInput) (*model.Site, string, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	domain := strings.TrimSpace(input.Domain)
	moderation := strings.TrimSpace(input.CommentModeration)
	if moderation == "" {
		moderation = model.CommentModerationPending
	}

	if err := validateSiteName(name); err != nil {
		return nil, "", err
	}
	if err := validateSlug(slug); err != nil {
		return nil, "", err
	}
	if err := validateDomain(domain); err != nil {
		return nil, "", err
	}
	if err := validateModeration(moderation); err != nil {
		return nil, "", err
	}

	exists, err := s.repository.SiteSlugExists(slug)
	if err != nil {
		return nil, "", fmt.Errorf("check site slug: %w", err)
	}
	if exists {
		return nil, "", ErrSlugExists
	}

	origins, err := normalizeOrigins(input.AllowedOrigins)
	if err != nil {
		return nil, "", err
	}
	key, prefix, hash, err := s.generateSiteKey()
	if err != nil {
		return nil, "", err
	}

	site := &model.Site{
		Name:              name,
		Slug:              slug,
		Domain:            domain,
		Status:            model.SiteStatusActive,
		SiteKeyPrefix:     prefix,
		SiteKeyHash:       hash,
		CommentModeration: moderation,
	}
	if err := site.SetAllowedOrigins(origins); err != nil {
		return nil, "", fmt.Errorf("encode allowed origins: %w", err)
	}
	if err := s.repository.CreateSite(site); err != nil {
		return nil, "", fmt.Errorf("create site: %w", err)
	}

	return site, key, nil
}

func (s *Service) ListSites(page, pageSize int, search string) ([]model.Site, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	sites, total, err := s.repository.ListSites(page, pageSize, strings.TrimSpace(search))
	if err != nil {
		return nil, 0, fmt.Errorf("list sites: %w", err)
	}
	return sites, total, nil
}

func (s *Service) GetSite(id uint) (*model.Site, error) {
	site, err := s.repository.FindSiteByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSiteNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find site: %w", err)
	}
	return site, nil
}

func (s *Service) UpdateSite(id uint, input UpdateSiteInput) (*model.Site, error) {
	if _, err := s.GetSite(id); err != nil {
		return nil, err
	}

	updates := make(map[string]any)
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if err := validateSiteName(name); err != nil {
			return nil, err
		}
		updates["name"] = name
	}
	if input.Domain != nil {
		domain := strings.TrimSpace(*input.Domain)
		if err := validateDomain(domain); err != nil {
			return nil, err
		}
		updates["domain"] = domain
	}
	if input.Status != nil {
		status := strings.TrimSpace(*input.Status)
		if err := validateStatus(status); err != nil {
			return nil, err
		}
		updates["status"] = status
	}
	if input.CommentModeration != nil {
		moderation := strings.TrimSpace(*input.CommentModeration)
		if err := validateModeration(moderation); err != nil {
			return nil, err
		}
		updates["comment_moderation"] = moderation
	}
	if input.AllowedOrigins != nil {
		origins, err := normalizeOrigins(*input.AllowedOrigins)
		if err != nil {
			return nil, err
		}
		encoded, err := json.Marshal(origins)
		if err != nil {
			return nil, fmt.Errorf("encode allowed origins: %w", err)
		}
		updates["allowed_origins"] = string(encoded)
	}
	if len(updates) == 0 {
		return nil, fmt.Errorf("%w: no fields to update", ErrInvalidSiteInput)
	}

	if err := s.repository.UpdateSite(id, updates); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSiteNotFound
		}
		return nil, fmt.Errorf("update site: %w", err)
	}
	return s.GetSite(id)
}

func (s *Service) DeleteSite(id uint) error {
	if err := s.repository.DeleteSite(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSiteNotFound
		}
		return fmt.Errorf("delete site: %w", err)
	}
	return nil
}

func (s *Service) RotateSiteKey(id uint) (*model.Site, string, error) {
	site, err := s.GetSite(id)
	if err != nil {
		return nil, "", err
	}

	key, prefix, hash, err := s.generateSiteKey()
	if err != nil {
		return nil, "", err
	}
	if err := s.repository.UpdateSiteKey(id, prefix, hash); err != nil {
		return nil, "", fmt.Errorf("rotate site key: %w", err)
	}
	site.SiteKeyPrefix = prefix
	site.SiteKeyHash = hash
	return site, key, nil
}

func (s *Service) RevokeSiteKey(id uint) error {
	if _, err := s.GetSite(id); err != nil {
		return err
	}
	if err := s.repository.UpdateSiteKey(id, "", ""); err != nil {
		return fmt.Errorf("revoke site key: %w", err)
	}
	return nil
}

func (s *Service) ResolveSiteByKey(key string) (*model.Site, error) {
	key = strings.TrimSpace(key)
	if !security.ValidSiteKey(key) {
		return nil, ErrInvalidSiteKey
	}

	site, err := s.repository.FindSiteByKeyPrefix(security.SiteKeyPrefix(key))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidSiteKey
	}
	if err != nil {
		return nil, fmt.Errorf("find site by key: %w", err)
	}
	if !security.VerifySiteKey(key, site.SiteKeyHash) {
		return nil, ErrInvalidSiteKey
	}
	if site.Status != model.SiteStatusActive {
		return nil, ErrSiteDisabled
	}
	return site, nil
}

func (s *Service) generateSiteKey() (string, string, string, error) {
	for attempt := 0; attempt < 3; attempt++ {
		key, prefix, hash, err := security.GenerateSiteKey()
		if err != nil {
			return "", "", "", err
		}
		if _, err := s.repository.FindSiteByKeyPrefix(prefix); errors.Is(err, gorm.ErrRecordNotFound) {
			return key, prefix, hash, nil
		} else if err != nil {
			return "", "", "", fmt.Errorf("check site key prefix: %w", err)
		}
	}
	return "", "", "", fmt.Errorf("generate unique site key: exhausted attempts")
}

func validateSiteName(name string) error {
	length := utf8.RuneCountInString(name)
	if length < 1 || length > 100 {
		return fmt.Errorf("%w: site name must be between 1 and 100 characters", ErrInvalidSiteInput)
	}
	return nil
}

func validateSlug(slug string) error {
	if len(slug) < 2 || len(slug) > 64 || !slugPattern.MatchString(slug) {
		return fmt.Errorf("%w: slug must be 2-64 lowercase letters, digits or hyphens", ErrInvalidSiteInput)
	}
	return nil
}

func validateDomain(domain string) error {
	if len(domain) > 255 {
		return fmt.Errorf("%w: domain must be at most 255 characters", ErrInvalidSiteInput)
	}
	return nil
}

func validateStatus(status string) error {
	if status != model.SiteStatusActive && status != model.SiteStatusDisabled {
		return fmt.Errorf("%w: status must be active or disabled", ErrInvalidSiteInput)
	}
	return nil
}

func validateModeration(moderation string) error {
	if moderation != model.CommentModerationPending && moderation != model.CommentModerationApproved {
		return fmt.Errorf("%w: commentModeration must be pending or approved", ErrInvalidSiteInput)
	}
	return nil
}

func normalizeOrigins(origins []string) ([]string, error) {
	normalized := make([]string, 0, len(origins))
	seen := make(map[string]struct{}, len(origins))

	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			return nil, fmt.Errorf("%w: wildcard origins are not allowed; leave the list empty to allow all", ErrInvalidSiteInput)
		}

		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return nil, fmt.Errorf("%w: invalid origin %q", ErrInvalidSiteInput, origin)
		}
		if parsed.Path != "" && parsed.Path != "/" {
			return nil, fmt.Errorf("%w: origin %q must not contain a path", ErrInvalidSiteInput, origin)
		}
		if parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("%w: origin %q must not contain query or fragment", ErrInvalidSiteInput, origin)
		}

		value := parsed.Scheme + "://" + parsed.Host
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}

	return normalized, nil
}
