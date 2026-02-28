package domain

import "time"

// Category maps to the Supabase public.category table.
type Category struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	ShortName string    `json:"short_name" db:"short_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Subcategory maps to the Supabase public.subcategory table.
type Subcategory struct {
	ID         int64     `json:"id" db:"id"`
	CategoryID int64     `json:"category_id" db:"category_id"`
	Name       string    `json:"name" db:"name"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
