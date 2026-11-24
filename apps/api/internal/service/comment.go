package service

import (
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"marku/api/internal/model"
	"marku/api/internal/repository"
	"marku/api/internal/security"
)

var (
	commentLinkPattern = regexp.MustCompile(`(?i)https?://`)
	validCommentStatus = map[string]struct{}{
		model.CommentStatusPending:  {},
		model.CommentStatusApproved: {},
		model.CommentStatusRejected: {},
		model.CommentStatusSpam:     {},
	}
)

type CreatePublicCommentInput struct {
	Path          string
	ParentID      *uint
	AuthorName    string
	AuthorEmail   string
	AuthorWebsite string
	Content       string
	Honeypot      string
	FormStartedAt int64
}

type CommentNode struct {
	Comment model.Comment
	Replies []CommentNode
}

type CommentPage struct {
	Items    []CommentNode
	Total    int64
	Page     int
	PageSize int
}

type CommentTrendPoint struct {
	Date     string `json:"date"`
	Comments int64  `json:"comments"`
	Pending  int64  `json:"pending"`
}

type SiteCommentStats struct {
	Total    int64               `json:"total"`
	Pending  int64               `json:"pending"`
	Approved int64               `json:"approved"`
	Trend    []CommentTrendPoint `json:"trend"`
}

type AdminCommentFilters struct {
	SiteID   uint
	Status   string
	Path     string
	Search   string
	From     *time.Time
	To       *time.Time
	Page     int
	PageSize int
	Sort     string
}

func (s *Service) GetSiteCommentStats(siteID uint, days int) (*SiteCommentStats, error) {
	if _, err := s.GetSite(siteID); err != nil {
		return nil, err
	}
	if days < 1 {
		days = 7
	}
	if days > 365 {
		days = 365
	}

	totals, err := s.repository.GetSiteCommentTotals(siteID)
	if err != nil {
		return nil, fmt.Errorf("get site comment totals: %w", err)
	}

	now := time.Now().UTC()
	since := now.AddDate(0, 0, -(days - 1))
	rows, err := s.repository.ListDailyCommentCounts(siteID, since)
	if err != nil {
		return nil, fmt.Errorf("list daily comment counts: %w", err)
	}
	byDate := make(map[string]CommentTrendPoint, len(rows))
	for _, row := range rows {
		byDate[row.Date] = CommentTrendPoint{
			Date:     row.Date,
			Comments: row.Comments,
			Pending:  row.Pending,
		}
	}

	trend := make([]CommentTrendPoint, 0, days)
	for offset := days - 1; offset >= 0; offset-- {
		date := now.AddDate(0, 0, -offset).Format("2006-01-02")
		if point, ok := byDate[date]; ok {
			trend = append(trend, point)
		} else {
			trend = append(trend, CommentTrendPoint{Date: date})
		}
	}

	return &SiteCommentStats{
		Total:    totals.Total,
		Pending:  totals.Pending,
		Approved: totals.Approved,
		Trend:    trend,
	}, nil
}

func (s *Service) CreatePublicComment(site *model.Site, input CreatePublicCommentInput, clientIP, userAgent string) (*model.Comment, error) {
	if site == nil {
		return nil, fmt.Errorf("%w: site is required", ErrInvalidCommentInput)
	}
	if site.Status != model.SiteStatusActive {
		return nil, ErrSiteDisabled
	}
	if strings.TrimSpace(input.Honeypot) != "" {
		return nil, fmt.Errorf("%w: invalid submission", ErrInvalidCommentInput)
	}

	now := time.Now().UTC()
	if input.FormStartedAt > 0 {
		startedAt := time.UnixMilli(input.FormStartedAt)
		elapsed := now.Sub(startedAt)
		if elapsed < 0 {
			return nil, fmt.Errorf("%w: invalid form timestamp", ErrInvalidCommentInput)
		}
		if elapsed < s.comment.MinSubmitDuration {
			return nil, ErrCommentTooFast
		}
		if elapsed > 24*time.Hour {
			return nil, fmt.Errorf("%w: form expired", ErrInvalidCommentInput)
		}
	}

	path, err := normalizePagePath(input.Path)
	if err != nil {
		return nil, err
	}
	name, err := normalizeAuthorName(input.AuthorName)
	if err != nil {
		return nil, err
	}
	email, err := normalizeAuthorEmail(input.AuthorEmail)
	if err != nil {
		return nil, err
	}
	website, err := normalizeAuthorWebsite(input.AuthorWebsite)
	if err != nil {
		return nil, err
	}
	content, err := normalizeCommentContent(input.Content, s.comment.MaxLength, s.comment.MaxLinks)
	if err != nil {
		return nil, err
	}

	ipHash := security.HashIP(clientIP, s.comment.IPHashSalt)
	if ipHash != "" {
		count, err := s.repository.CountCommentsByIP(site.ID, ipHash, now.Add(-s.comment.RateLimitWindow))
		if err != nil {
			return nil, fmt.Errorf("count comments by ip: %w", err)
		}
		if count >= int64(s.comment.RateLimitMax) {
			return nil, ErrCommentRateLimited
		}
	}

	contentHash := security.HashContent(content)
	duplicate, err := s.repository.CommentExistsSince(
		site.ID,
		path,
		ipHash,
		contentHash,
		now.Add(-s.comment.DuplicateWindow),
	)
	if err != nil {
		return nil, fmt.Errorf("check duplicate comment: %w", err)
	}
	if duplicate {
		return nil, ErrDuplicateComment
	}

	if input.ParentID != nil {
		parent, err := s.repository.FindCommentByIDAndSite(*input.ParentID, site.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidParent
		}
		if err != nil {
			return nil, fmt.Errorf("find parent comment: %w", err)
		}
		if parent.Path != path || parent.Status != model.CommentStatusApproved {
			return nil, ErrInvalidParent
		}
		depth, err := s.commentDepth(parent, site.ID)
		if err != nil {
			return nil, err
		}
		if depth >= s.comment.MaxDepth {
			return nil, ErrCommentDepthExceeded
		}
	}

	status := model.CommentStatusPending
	if site.CommentModeration == model.CommentModerationApproved {
		status = model.CommentStatusApproved
	}

	comment := &model.Comment{
		SiteID:        site.ID,
		Path:          path,
		ParentID:      input.ParentID,
		AuthorName:    name,
		AuthorEmail:   email,
		AuthorWebsite: website,
		Content:       content,
		ContentHash:   contentHash,
		Status:        status,
		IsAdmin:       false,
		IPHash:        ipHash,
		UserAgent:     truncate(userAgent, 512),
	}
	if err := s.repository.CreateComment(comment); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	return comment, nil
}

func (s *Service) ListPublicComments(siteID uint, rawPath string, page, pageSize int, sortOrder string) (CommentPage, error) {
	path, err := normalizePagePath(rawPath)
	if err != nil {
		return CommentPage{}, err
	}
	page, pageSize = normalizePage(page, pageSize)
	if sortOrder != "desc" {
		sortOrder = "asc"
	}

	comments, err := s.repository.ListApprovedCommentsByPath(siteID, path)
	if err != nil {
		return CommentPage{}, fmt.Errorf("list approved comments: %w", err)
	}

	roots := buildCommentTree(comments, s.comment.MaxDepth, sortOrder)
	total := int64(len(roots))
	start := (page - 1) * pageSize
	if start > len(roots) {
		start = len(roots)
	}
	end := start + pageSize
	if end > len(roots) {
		end = len(roots)
	}

	return CommentPage{
		Items:    roots[start:end],
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *Service) CountApprovedComments(siteID uint, rawPath string) (int64, error) {
	path, err := normalizePagePath(rawPath)
	if err != nil {
		return 0, err
	}
	comments, err := s.repository.ListApprovedCommentsByPath(siteID, path)
	if err != nil {
		return 0, fmt.Errorf("list approved comments: %w", err)
	}
	roots := buildCommentTree(comments, s.comment.MaxDepth, "asc")
	return countCommentNodes(roots), nil
}

func (s *Service) ListAdminComments(filters AdminCommentFilters) ([]model.Comment, int64, error) {
	if _, err := s.GetSite(filters.SiteID); err != nil {
		return nil, 0, err
	}

	status := strings.TrimSpace(filters.Status)
	if status != "" {
		if err := validateCommentStatus(status); err != nil {
			return nil, 0, err
		}
	}

	path := ""
	if strings.TrimSpace(filters.Path) != "" {
		normalized, err := normalizePagePath(filters.Path)
		if err != nil {
			return nil, 0, err
		}
		path = normalized
	}

	page, pageSize := normalizePage(filters.Page, filters.PageSize)
	sortOrder := "desc"
	if filters.Sort == "asc" {
		sortOrder = "asc"
	}

	comments, total, err := s.repository.ListComments(repository.CommentListFilter{
		SiteID:   filters.SiteID,
		Status:   status,
		Path:     path,
		Search:   strings.TrimSpace(filters.Search),
		From:     filters.From,
		To:       filters.To,
		Page:     page,
		PageSize: pageSize,
		Sort:     sortOrder,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list comments: %w", err)
	}
	return comments, total, nil
}

func (s *Service) UpdateCommentStatus(id uint, status string) (*model.Comment, error) {
	if err := validateCommentStatus(status); err != nil {
		return nil, err
	}
	updated, err := s.repository.UpdateCommentStatus(id, status)
	if err != nil {
		return nil, fmt.Errorf("update comment status: %w", err)
	}
	if updated == 0 {
		return nil, ErrCommentNotFound
	}
	comment, err := s.repository.FindCommentByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCommentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find comment: %w", err)
	}
	return comment, nil
}

func (s *Service) DeleteComment(id uint) error {
	deleted, err := s.repository.DeleteComment(id)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	if deleted == 0 {
		return ErrCommentNotFound
	}
	return nil
}

func (s *Service) ReplyComment(commentID uint, admin *model.Admin, content string) (*model.Comment, error) {
	if admin == nil {
		return nil, ErrInvalidToken
	}

	parent, err := s.repository.FindCommentByID(commentID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCommentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find parent comment: %w", err)
	}

	normalizedContent, err := normalizeCommentContent(content, s.comment.MaxLength, s.comment.MaxLinks)
	if err != nil {
		return nil, err
	}
	depth, err := s.commentDepth(parent, parent.SiteID)
	if err != nil {
		return nil, err
	}
	if depth >= s.comment.MaxDepth {
		return nil, ErrCommentDepthExceeded
	}

	nickname := strings.TrimSpace(admin.Nickname)
	if nickname == "" {
		nickname = "Admin"
	}
	reply := &model.Comment{
		SiteID:        parent.SiteID,
		Path:          parent.Path,
		ParentID:      &commentID,
		AuthorName:    nickname,
		AuthorEmail:   admin.Email,
		AuthorWebsite: "",
		Content:       normalizedContent,
		ContentHash:   security.HashContent(normalizedContent),
		Status:        model.CommentStatusApproved,
		IsAdmin:       true,
		UserAgent:     "admin",
	}
	if err := s.repository.CreateComment(reply); err != nil {
		return nil, fmt.Errorf("create admin reply: %w", err)
	}
	return reply, nil
}

func (s *Service) BulkCommentAction(siteID uint, ids []uint, action string) (int64, error) {
	if _, err := s.GetSite(siteID); err != nil {
		return 0, err
	}
	if len(ids) == 0 || len(ids) > 100 {
		return 0, fmt.Errorf("%w: ids must contain between 1 and 100 items", ErrInvalidCommentInput)
	}

	switch action {
	case "approve", "approved":
		return s.repository.UpdateCommentsStatusInSite(siteID, ids, model.CommentStatusApproved)
	case "reject", "rejected":
		return s.repository.UpdateCommentsStatusInSite(siteID, ids, model.CommentStatusRejected)
	case "spam":
		return s.repository.UpdateCommentsStatusInSite(siteID, ids, model.CommentStatusSpam)
	case "delete":
		return s.repository.DeleteCommentsInSite(siteID, ids)
	default:
		return 0, ErrInvalidCommentAction
	}
}

func buildCommentTree(comments []model.Comment, maxDepth int, sortOrder string) []CommentNode {
	children := make(map[uint][]*model.Comment)
	roots := make([]*model.Comment, 0)
	for index := range comments {
		comment := &comments[index]
		if comment.ParentID == nil {
			roots = append(roots, comment)
			continue
		}
		children[*comment.ParentID] = append(children[*comment.ParentID], comment)
	}

	sort.SliceStable(roots, func(i, j int) bool {
		if sortOrder == "desc" {
			return roots[i].CreatedAt.After(roots[j].CreatedAt)
		}
		return roots[i].CreatedAt.Before(roots[j].CreatedAt)
	})

	var build func(comment *model.Comment, depth int) CommentNode
	build = func(comment *model.Comment, depth int) CommentNode {
		node := CommentNode{Comment: *comment}
		if depth >= maxDepth {
			return node
		}
		for _, child := range children[comment.ID] {
			node.Replies = append(node.Replies, build(child, depth+1))
		}
		return node
	}

	nodes := make([]CommentNode, 0, len(roots))
	for _, root := range roots {
		nodes = append(nodes, build(root, 1))
	}
	return nodes
}

func countCommentNodes(nodes []CommentNode) int64 {
	var count int64
	for _, node := range nodes {
		count++
		count += countCommentNodes(node.Replies)
	}
	return count
}

func (s *Service) commentDepth(comment *model.Comment, siteID uint) (int, error) {
	depth := 1
	current := comment
	for current.ParentID != nil {
		if depth > 100 {
			return 0, fmt.Errorf("%w: comment chain too deep", ErrInvalidParent)
		}
		parent, err := s.repository.FindCommentByIDAndSite(*current.ParentID, siteID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrInvalidParent
		}
		if err != nil {
			return 0, fmt.Errorf("find parent comment: %w", err)
		}
		depth++
		current = parent
	}
	return depth, nil
}

func normalizeAuthorName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	length := utf8.RuneCountInString(name)
	if length < 1 || length > 100 {
		return "", fmt.Errorf("%w: author name must be between 1 and 100 characters", ErrInvalidCommentInput)
	}
	return name, nil
}

func normalizeAuthorEmail(raw string) (string, error) {
	email := strings.TrimSpace(raw)
	if email == "" {
		return "", nil
	}
	if len(email) > 255 {
		return "", fmt.Errorf("%w: author email is too long", ErrInvalidCommentInput)
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", fmt.Errorf("%w: author email is invalid", ErrInvalidCommentInput)
	}
	return email, nil
}

func normalizeAuthorWebsite(raw string) (string, error) {
	website := strings.TrimSpace(raw)
	if website == "" {
		return "", nil
	}
	if len(website) > 255 {
		return "", fmt.Errorf("%w: author website is too long", ErrInvalidCommentInput)
	}
	parsed, err := url.Parse(website)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", fmt.Errorf("%w: author website must be a valid http(s) URL", ErrInvalidCommentInput)
	}
	return website, nil
}

func normalizeCommentContent(raw string, maxLength, maxLinks int) (string, error) {
	content := strings.ReplaceAll(raw, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	content = strings.TrimSpace(content)
	if content == "" {
		return "", fmt.Errorf("%w: content is required", ErrInvalidCommentInput)
	}
	if utf8.RuneCountInString(content) > maxLength {
		return "", fmt.Errorf("%w: content must be at most %d characters", ErrInvalidCommentInput, maxLength)
	}
	for _, char := range content {
		if char == 0 || (char < 32 && char != '\n' && char != '\t') {
			return "", fmt.Errorf("%w: content contains invalid characters", ErrInvalidCommentInput)
		}
	}
	if len(commentLinkPattern.FindAllString(content, -1)) > maxLinks {
		return "", fmt.Errorf("%w: content contains too many links", ErrInvalidCommentInput)
	}
	return content, nil
}

func validateCommentStatus(status string) error {
	if _, ok := validCommentStatus[status]; !ok {
		return fmt.Errorf("%w: status must be pending, approved, rejected or spam", ErrInvalidCommentInput)
	}
	return nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func truncate(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	return value[:maxBytes]
}
