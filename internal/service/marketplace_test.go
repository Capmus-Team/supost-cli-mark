package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Capmus-Team/supost-cli-mark/internal/domain"
)

type mockMarketplaceRepo struct {
	filter        domain.PostFilter
	categories    []domain.Category
	subcategories []domain.Subcategory
	posts         []domain.Post
	total         int
}

func (m *mockMarketplaceRepo) ListCategories(_ context.Context) ([]domain.Category, error) {
	return m.categories, nil
}

func (m *mockMarketplaceRepo) ListSubcategoriesByCategory(_ context.Context, _ int64) ([]domain.Subcategory, error) {
	return m.subcategories, nil
}

func (m *mockMarketplaceRepo) ListPosts(_ context.Context, filter domain.PostFilter) ([]domain.Post, int, error) {
	m.filter = filter
	return m.posts, m.total, nil
}

func TestMarketplaceService_ListSubcategories_Validation(t *testing.T) {
	repo := &mockMarketplaceRepo{}
	svc := NewMarketplaceService(repo)

	_, err := svc.ListSubcategories(context.Background(), 0)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestMarketplaceService_ListPosts_DefaultsAndClamp(t *testing.T) {
	repo := &mockMarketplaceRepo{}
	svc := NewMarketplaceService(repo)

	_, _, err := svc.ListPosts(context.Background(), domain.PostFilter{
		Limit:  999,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.filter.Limit != maxPostsLimit {
		t.Fatalf("expected clamped limit %d, got %d", maxPostsLimit, repo.filter.Limit)
	}
	if repo.filter.Status != domain.PostStatusActive {
		t.Fatalf("expected default status %d, got %d", domain.PostStatusActive, repo.filter.Status)
	}
}

func TestMarketplaceService_ListPosts_ValidatesNegativeValues(t *testing.T) {
	repo := &mockMarketplaceRepo{}
	svc := NewMarketplaceService(repo)

	_, _, err := svc.ListPosts(context.Background(), domain.PostFilter{Offset: -1})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}
