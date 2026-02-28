package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/Capmus-Team/supost-cli-mark/internal/domain"
)

// InMemory implements ListingStore using an in-memory map.
// Perfect for prototyping and testing — zero external dependencies.
// Swap to Postgres when ready. See AGENTS.md §6.5.
type InMemory struct {
	mu       sync.RWMutex
	listings map[string]domain.Listing
	posts    []domain.Post
	cats     []domain.Category
	subcats  []domain.Subcategory
}

// NewInMemory creates a new in-memory repository pre-loaded with seed data.
func NewInMemory() *InMemory {
	repo := &InMemory{
		listings: make(map[string]domain.Listing),
		posts:    make([]domain.Post, 0),
		cats:     make([]domain.Category, 0),
		subcats:  make([]domain.Subcategory, 0),
	}
	repo.loadSeedData()
	repo.loadPostSeedData()
	repo.loadCategorySeedData()
	repo.loadSubcategorySeedData()
	return repo
}

func (r *InMemory) ListActive(_ context.Context) ([]domain.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var active []domain.Listing
	for _, l := range r.listings {
		if l.Status == "active" {
			active = append(active, l)
		}
	}
	return active, nil
}

func (r *InMemory) GetByID(_ context.Context, id string) (*domain.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	l, ok := r.listings[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &l, nil
}

func (r *InMemory) Create(_ context.Context, listing *domain.Listing) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if listing.ID == "" {
		listing.ID = fmt.Sprintf("mem-%d", len(r.listings)+1)
	}
	now := time.Now()
	listing.CreatedAt = now
	listing.UpdatedAt = now
	if listing.Status == "" {
		listing.Status = "active"
	}

	r.listings[listing.ID] = *listing
	return nil
}

func (r *InMemory) ListRecentActivePosts(_ context.Context, limit int) ([]domain.Post, error) {
	return r.listRecentActivePosts(limit, nil), nil
}

func (r *InMemory) ListRecentActivePostsByCategory(_ context.Context, categoryID int64, limit int) ([]domain.Post, error) {
	return r.listRecentActivePosts(limit, &categoryID), nil
}

func (r *InMemory) ListCategories(_ context.Context) ([]domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]domain.Category, len(r.cats))
	copy(out, r.cats)
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (r *InMemory) ListSubcategoriesByCategory(_ context.Context, categoryID int64) ([]domain.Subcategory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]domain.Subcategory, 0, len(r.subcats))
	for _, subcat := range r.subcats {
		if subcat.CategoryID == categoryID {
			out = append(out, subcat)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (r *InMemory) ListPosts(_ context.Context, filter domain.PostFilter) ([]domain.Post, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	filtered := make([]domain.Post, 0, len(r.posts))
	for _, post := range r.posts {
		if filter.CategoryID > 0 && post.CategoryID != filter.CategoryID {
			continue
		}
		if filter.SubcategoryID > 0 && post.SubcategoryID != filter.SubcategoryID {
			continue
		}
		if filter.Status != 0 && post.Status != filter.Status {
			continue
		}
		filtered = append(filtered, post)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].TimePosted == filtered[j].TimePosted {
			return filtered[i].ID > filtered[j].ID
		}
		return filtered[i].TimePosted > filtered[j].TimePosted
	})

	total := len(filtered)
	start := filter.Offset
	if start > total {
		start = total
	}
	end := start + filter.Limit
	if filter.Limit <= 0 || end > total {
		end = total
	}

	out := make([]domain.Post, 0, end-start)
	out = append(out, filtered[start:end]...)
	return out, total, nil
}

func (r *InMemory) listRecentActivePosts(limit int, categoryID *int64) []domain.Post {
	if limit <= 0 {
		limit = 50
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	active := make([]domain.Post, 0, len(r.posts))
	for _, post := range r.posts {
		if post.Status != domain.PostStatusActive {
			continue
		}
		if categoryID != nil && post.CategoryID != *categoryID {
			continue
		}
		active = append(active, post)
	}

	sort.Slice(active, func(i, j int) bool {
		if active[i].TimePosted == active[j].TimePosted {
			return active[i].ID > active[j].ID
		}
		return active[i].TimePosted > active[j].TimePosted
	})

	if len(active) > limit {
		active = active[:limit]
	}

	out := make([]domain.Post, len(active))
	copy(out, active)
	return out
}

// ListHomeCategorySections returns latest active post times per category.
func (r *InMemory) ListHomeCategorySections(_ context.Context) ([]domain.HomeCategorySection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	latestByCategory := make(map[int64]time.Time, 8)
	for _, post := range r.posts {
		if post.Status != domain.PostStatusActive {
			continue
		}
		var postedAt time.Time
		if !post.TimePostedAt.IsZero() {
			postedAt = post.TimePostedAt
		} else if post.TimePosted > 0 {
			postedAt = time.Unix(post.TimePosted, 0)
		}
		if postedAt.IsZero() {
			continue
		}
		if existing, ok := latestByCategory[post.CategoryID]; !ok || postedAt.After(existing) {
			latestByCategory[post.CategoryID] = postedAt
		}
	}

	sections := make([]domain.HomeCategorySection, 0, len(latestByCategory))
	for categoryID, postedAt := range latestByCategory {
		sections = append(sections, domain.HomeCategorySection{
			CategoryID:   categoryID,
			LastPostedAt: postedAt,
		})
	}
	sort.Slice(sections, func(i, j int) bool {
		return sections[i].CategoryID < sections[j].CategoryID
	})
	return sections, nil
}

// loadSeedData populates the repository with sample data.
// In a more advanced setup, this could read from testdata/seed/*.json.
func (r *InMemory) loadSeedData() {
	now := time.Now()
	seeds := []domain.Listing{
		{
			ID:          "seed-1",
			UserID:      "user-1",
			Title:       "Used Calculus Textbook",
			Description: "Stewart Calculus, 8th edition. Some highlighting.",
			Price:       4500,
			Category:    "textbooks",
			Status:      "active",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "seed-2",
			UserID:      "user-1",
			Title:       "IKEA Desk",
			Description: "MALM desk, white. Good condition. Pickup only.",
			Price:       6000,
			Category:    "furniture",
			Status:      "active",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "seed-3",
			UserID:      "user-2",
			Title:       "Trek Road Bike",
			Description: "2021 Domane AL 2, 56cm. Low miles.",
			Price:       55000,
			Category:    "bikes",
			Status:      "active",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	for _, s := range seeds {
		r.listings[s.ID] = s
	}
}

func (r *InMemory) loadPostSeedData() {
	now := time.Now()
	r.posts = append(r.posts,
		domain.Post{
			ID:            130031901,
			CategoryID:    3,
			SubcategoryID: 14,
			Email:         "alex@stanford.edu",
			Name:          "Sublet room in EVGR premium 2b2b",
			Status:        domain.PostStatusActive,
			TimePosted:    now.Add(-2 * time.Hour).Unix(),
			TimePostedAt:  now.Add(-2 * time.Hour),
			Price:         2000,
			HasPrice:      true,
			HasImage:      false,
			CreatedAt:     now.Add(-2 * time.Hour),
			UpdatedAt:     now.Add(-2 * time.Hour),
		},
		domain.Post{
			ID:            130031900,
			CategoryID:    3,
			SubcategoryID: 14,
			Email:         "casey@stanford.edu",
			Name:          "Shared House",
			Status:        domain.PostStatusActive,
			TimePosted:    now.Add(-3 * time.Hour).Unix(),
			TimePostedAt:  now.Add(-3 * time.Hour),
			Price:         700,
			HasPrice:      true,
			HasImage:      true,
			CreatedAt:     now.Add(-3 * time.Hour),
			UpdatedAt:     now.Add(-3 * time.Hour),
		},
		domain.Post{
			ID:            130031899,
			CategoryID:    5,
			SubcategoryID: 20,
			Email:         "morgan@stanford.edu",
			Name:          "Ikea Stackable beds(2) + 2 Mattresses - Pickup in MV FRIDAY / SATURDAY",
			Status:        domain.PostStatusActive,
			TimePosted:    now.Add(-5 * time.Hour).Unix(),
			TimePostedAt:  now.Add(-5 * time.Hour),
			Price:         0,
			HasPrice:      true,
			HasImage:      true,
			CreatedAt:     now.Add(-5 * time.Hour),
			UpdatedAt:     now.Add(-5 * time.Hour),
		},
		domain.Post{
			ID:            130031898,
			CategoryID:    9,
			SubcategoryID: 90,
			Email:         "sam@stanford.edu",
			Name:          "English tutoring (book club)",
			Status:        domain.PostStatusActive,
			TimePosted:    now.Add(-6 * time.Hour).Unix(),
			TimePostedAt:  now.Add(-6 * time.Hour),
			HasPrice:      false,
			HasImage:      false,
			CreatedAt:     now.Add(-6 * time.Hour),
			UpdatedAt:     now.Add(-6 * time.Hour),
		},
		domain.Post{
			ID:            130031897,
			CategoryID:    5,
			SubcategoryID: 20,
			Email:         "jamie@stanford.edu",
			Name:          "Table and chair",
			Status:        domain.PostStatusActive,
			TimePosted:    now.Add(-15 * time.Hour).Unix(),
			TimePostedAt:  now.Add(-15 * time.Hour),
			Price:         0,
			HasPrice:      true,
			HasImage:      true,
			CreatedAt:     now.Add(-15 * time.Hour),
			UpdatedAt:     now.Add(-15 * time.Hour),
		},
		domain.Post{
			ID:            130031896,
			CategoryID:    5,
			SubcategoryID: 20,
			Email:         "pat@stanford.edu",
			Name:          "Apple Magic Keyboard & Magic Mouse 2 Bundle (Lightning)",
			Status:        0,
			TimePosted:    now.Add(-20 * time.Hour).Unix(),
			TimePostedAt:  now.Add(-20 * time.Hour),
			Price:         65,
			HasPrice:      true,
			HasImage:      true,
			CreatedAt:     now.Add(-20 * time.Hour),
			UpdatedAt:     now.Add(-20 * time.Hour),
		},
	)
}

func (r *InMemory) loadCategorySeedData() {
	type categorySeedRow struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		ShortName string `json:"short_name"`
	}

	rows := make([]categorySeedRow, 0)
	if err := loadJSONSeed(filepath.Join("testdata", "seed", "category_rows.json"), &rows); err != nil {
		return
	}

	now := time.Now().UTC()
	for _, row := range rows {
		r.cats = append(r.cats, domain.Category{
			ID:        row.ID,
			Name:      row.Name,
			ShortName: row.ShortName,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
}

func (r *InMemory) loadSubcategorySeedData() {
	type subcategorySeedRow struct {
		ID         int64  `json:"id"`
		CategoryID int64  `json:"category_id"`
		Name       string `json:"name"`
	}

	rows := make([]subcategorySeedRow, 0)
	if err := loadJSONSeed(filepath.Join("testdata", "seed", "subcategory_rows.json"), &rows); err != nil {
		return
	}

	now := time.Now().UTC()
	for _, row := range rows {
		r.subcats = append(r.subcats, domain.Subcategory{
			ID:         row.ID,
			CategoryID: row.CategoryID,
			Name:       row.Name,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
}

func loadJSONSeed(path string, target interface{}) error {
	candidates := []string{
		path,
		filepath.Join("..", "..", path),
	}

	var payload []byte
	var err error
	for _, candidate := range candidates {
		payload, err = os.ReadFile(candidate)
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}

	if err := json.Unmarshal(payload, target); err != nil {
		return err
	}
	return nil
}
