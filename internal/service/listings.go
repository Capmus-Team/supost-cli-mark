// Package service contains the core business logic.
// Services are 100% CLI-agnostic: no Cobra, no os.Exit(), no stdout.
// Services accept interfaces and are fully testable.
// See AGENTS.md §2.4, §6.1.
package service

import (
	"context"

	"github.com/Capmus-Team/supost-cli-mark/internal/domain"
)

// ListingRepository defines the data access interface for listings.
// Defined here (where consumed), not in repository/ (where implemented).
type ListingRepository interface {
	ListActive(ctx context.Context) ([]domain.Listing, error)
	GetByID(ctx context.Context, id string) (*domain.Listing, error)
	Create(ctx context.Context, listing *domain.Listing) error
}

// ListingService orchestrates listing-related business logic.
type ListingService struct {
	repo ListingRepository
}

// NewListingService creates a new ListingService.
func NewListingService(repo ListingRepository) *ListingService {
	return &ListingService{repo: repo}
}

// ListActive returns all active listings.
func (s *ListingService) ListActive(ctx context.Context) ([]domain.Listing, error) {
	return s.repo.ListActive(ctx)
}

// Create validates and persists a new listing.
func (s *ListingService) Create(ctx context.Context, listing *domain.Listing) error {
	if err := listing.Validate(); err != nil {
		return err
	}
	return s.repo.Create(ctx, listing)
}
