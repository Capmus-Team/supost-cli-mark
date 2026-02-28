package domain

// PostFilter is the filter/pagination input for listing posts.
type PostFilter struct {
	CategoryID    int64 `json:"category_id" db:"category_id"`
	SubcategoryID int64 `json:"subcategory_id" db:"subcategory_id"`
	Status        int   `json:"status" db:"status"`
	Limit         int   `json:"limit" db:"limit"`
	Offset        int   `json:"offset" db:"offset"`
}
