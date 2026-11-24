package service

import (
	"marku/api/internal/config"
	"marku/api/internal/repository"
	"marku/api/internal/security"
)

type Service struct {
	repository *repository.Store
	tokens     *security.TokenManager
	auth       config.AuthConfig
	comment    config.CommentConfig
	view       config.ViewConfig
}

func New(store *repository.Store, cfg config.Config) *Service {
	return &Service{
		repository: store,
		tokens:     security.NewTokenManager(cfg.Auth.JWTSecret, cfg.Auth.TokenTTL),
		auth:       cfg.Auth,
		comment:    cfg.Comment,
		view:       cfg.View,
	}
}

func (s *Service) Store() *repository.Store {
	return s.repository
}

func (s *Service) Tokens() *security.TokenManager {
	return s.tokens
}
