package service

import (
	"context"

	"github.com/Capmus-Team/supost-cli-mark/internal/domain"
)

const defaultHomeLimit = 50

// HomeRepository defines data access required by the home page.
type HomeRepository interface {
	ListRecentActivePosts(ctx context.Context, limit int) ([]domain.Post, error)
	ListRecentActivePostsByCategory(ctx context.Context, categoryID int64, limit int) ([]domain.Post, error)
	ListHomeCategorySections(ctx context.Context) ([]domain.HomeCategorySection, error)
}

// HomeService orchestrates homepage post retrieval.
type HomeService struct {
	repo HomeRepository
}

// NewHomeService constructs HomeService.
func NewHomeService(repo HomeRepository) *HomeService {
	return &HomeService{repo: repo}
}

// ListRecentActive returns the most recent active posts for home.
func (s *HomeService) ListRecentActive(ctx context.Context, limit int) ([]domain.Post, error) {
	if limit <= 0 {
		limit = defaultHomeLimit
	}
	return s.repo.ListRecentActivePosts(ctx, limit)
}

// ListRecentActiveByCategory returns the most recent active posts for one category.
func (s *HomeService) ListRecentActiveByCategory(ctx context.Context, categoryID int64, limit int) ([]domain.Post, error) {
	if limit <= 0 {
		limit = defaultHomeLimit
	}
	return s.repo.ListRecentActivePostsByCategory(ctx, categoryID, limit)
}

// ListCategorySections returns home sidebar category/subcategory sections.
func (s *HomeService) ListCategorySections(ctx context.Context) ([]domain.HomeCategorySection, error) {
	return s.repo.ListHomeCategorySections(ctx)
}
