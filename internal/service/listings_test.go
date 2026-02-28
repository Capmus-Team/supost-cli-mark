package service

import (
	"context"
	"testing"

	"github.com/Capmus-Team/supost-cli-mark/internal/domain"
)

// mockListingRepo is a minimal test double.
type mockListingRepo struct {
	listings []domain.Listing
	created  []*domain.Listing
}

func (m *mockListingRepo) ListActive(_ context.Context) ([]domain.Listing, error) {
	var active []domain.Listing
	for _, l := range m.listings {
		if l.Status == "active" {
			active = append(active, l)
		}
	}
	return active, nil
}

func (m *mockListingRepo) GetByID(_ context.Context, id string) (*domain.Listing, error) {
	for _, l := range m.listings {
		if l.ID == id {
			return &l, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockListingRepo) Create(_ context.Context, listing *domain.Listing) error {
	m.created = append(m.created, listing)
	return nil
}

func TestListingService_ListActive(t *testing.T) {
	repo := &mockListingRepo{
		listings: []domain.Listing{
			{ID: "1", Title: "Active Item", Status: "active"},
			{ID: "2", Title: "Sold Item", Status: "sold"},
		},
	}
	svc := NewListingService(repo)

	items, err := svc.ListActive(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 active listing, got %d", len(items))
	}
}

func TestListingService_Create_Validation(t *testing.T) {
	repo := &mockListingRepo{}
	svc := NewListingService(repo)

	tests := []struct {
		name    string
		listing domain.Listing
		wantErr bool
	}{
		{
			name:    "valid listing",
			listing: domain.Listing{Title: "Textbook", Price: 4500},
			wantErr: false,
		},
		{
			name:    "missing title",
			listing: domain.Listing{Title: "", Price: 4500},
			wantErr: true,
		},
		{
			name:    "negative price",
			listing: domain.Listing{Title: "Textbook", Price: -100},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(context.Background(), &tt.listing)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
