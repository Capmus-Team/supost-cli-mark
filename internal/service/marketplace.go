package service

import (
	"context"
	"fmt"

	"github.com/Capmus-Team/supost-cli-mark/internal/adapters"
	"github.com/Capmus-Team/supost-cli-mark/internal/domain"
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
	GetPostByID(ctx context.Context, id int64) (*domain.Post, error)
	SendMessage(ctx context.Context, postID int64, email, message string) error
}

// MarketplaceService validates inputs and orchestrates API reads.
type MarketplaceService struct {
	repo        MarketplaceRepository
	emailSender adapters.EmailSender
}

// NewMarketplaceService creates a new MarketplaceService.
// emailSender may be nil; if set, SendMessage will email the poster via Mailgun.
func NewMarketplaceService(repo MarketplaceRepository, emailSender adapters.EmailSender) *MarketplaceService {
	return &MarketplaceService{repo: repo, emailSender: emailSender}
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

// GetPost returns a single post by ID.
func (s *MarketplaceService) GetPost(ctx context.Context, id int64) (*domain.Post, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: post id must be greater than 0", domain.ErrValidation)
	}
	return s.repo.GetPostByID(ctx, id)
}

// SendMessage records a message and emails the post's seller (if Mailgun configured).
func (s *MarketplaceService) SendMessage(ctx context.Context, postID int64, email, message string) error {
	if postID <= 0 {
		return fmt.Errorf("%w: post id must be greater than 0", domain.ErrValidation)
	}
	if email == "" {
		return fmt.Errorf("%w: email is required", domain.ErrValidation)
	}
	if message == "" {
		return fmt.Errorf("%w: message is required", domain.ErrValidation)
	}

	if err := s.repo.SendMessage(ctx, postID, email, message); err != nil {
		return err
	}

	if s.emailSender != nil {
		post, err := s.repo.GetPostByID(ctx, postID)
		if err != nil || post == nil {
			return nil // persist succeeded; email is best-effort
		}
		posterEmail := post.Email
		if posterEmail == "" {
			return nil
		}
		if err := s.emailSender.SendPostMessage(ctx, posterEmail, email, post.Name, message); err != nil {
			return fmt.Errorf("sending email to poster: %w", err)
		}
	}
	return nil
}
