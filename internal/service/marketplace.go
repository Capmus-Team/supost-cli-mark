package service

import (
	"context"
	"fmt"

	"github.com/Capmus-Team/supost-cli/internal/domain"
)

const (
	defaultPostsLimit = 20
	maxPostsLimit     = 100
)

// MarketplaceRepository defines data access required by API endpoints.
type MarketplaceRepository interface {
	ListCategories(ctx context.Context) ([]domain.Category, error)
	ListSubcategoriesByCategory(ctx context.Context, categoryID int64) ([]domain.Subcategory, error)
	ListPosts(ctx context.Context, filter domain.PostFilter) ([]domain.Post, int, error)
}

// MarketplaceService validates inputs and orchestrates API reads.
type MarketplaceService struct {
	repo MarketplaceRepository
}

// NewMarketplaceService creates a new MarketplaceService.
func NewMarketplaceService(repo MarketplaceRepository) *MarketplaceService {
	return &MarketplaceService{repo: repo}
}

// ListCategories returns all categories.
func (s *MarketplaceService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.repo.ListCategories(ctx)
}

// ListSubcategories returns subcategories for a category id.
func (s *MarketplaceService) ListSubcategories(ctx context.Context, categoryID int64) ([]domain.Subcategory, error) {
	if categoryID <= 0 {
		return nil, fmt.Errorf("%w: category_id must be greater than 0", domain.ErrValidation)
	}
	return s.repo.ListSubcategoriesByCategory(ctx, categoryID)
}

// ListPosts returns filtered posts with pagination metadata.
func (s *MarketplaceService) ListPosts(ctx context.Context, filter domain.PostFilter) ([]domain.Post, int, error) {
	if filter.CategoryID < 0 {
		return nil, 0, fmt.Errorf("%w: category_id must be non-negative", domain.ErrValidation)
	}
	if filter.SubcategoryID < 0 {
		return nil, 0, fmt.Errorf("%w: subcategory_id must be non-negative", domain.ErrValidation)
	}
	if filter.Offset < 0 {
		return nil, 0, fmt.Errorf("%w: offset must be non-negative", domain.ErrValidation)
	}
	if filter.Limit <= 0 {
		filter.Limit = defaultPostsLimit
	}
	if filter.Limit > maxPostsLimit {
		filter.Limit = maxPostsLimit
	}
	if filter.Status == 0 {
		filter.Status = domain.PostStatusActive
	}

	return s.repo.ListPosts(ctx, filter)
}
