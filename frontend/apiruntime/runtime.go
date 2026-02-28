package apiruntime

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Runtime struct {
	db          *sql.DB
	corsOrigins map[string]struct{}
}

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	ShortName string    `json:"short_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Subcategory struct {
	ID         int64     `json:"id"`
	CategoryID int64     `json:"category_id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Post struct {
	ID            int64     `json:"id"`
	CategoryID    int64     `json:"category_id"`
	SubcategoryID int64     `json:"subcategory_id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	Body          string    `json:"body"`
	Status        int       `json:"status"`
	TimePosted    int64     `json:"time_posted"`
	TimePostedAt  time.Time `json:"time_posted_at"`
	Price         float64   `json:"price"`
	HasPrice      bool      `json:"has_price"`
	HasImage      bool      `json:"has_image"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

var (
	runtimeOnce sync.Once
	runtimeInst *Runtime
	runtimeErr  error
)

func GetRuntime() (*Runtime, error) {
	runtimeOnce.Do(func() {
		runtimeInst, runtimeErr = newRuntime()
	})
	return runtimeInst, runtimeErr
}

func newRuntime() (*Runtime, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required for Vercel backend")
	}

	connString := ensurePoolerSafeConnectionString(databaseURL)
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("opening postgres connection: %w", err)
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &Runtime{
		db:          db,
		corsOrigins: buildAllowedOrigins(os.Getenv("CORS_ORIGINS")),
	}, nil
}

func (rt *Runtime) Health(w http.ResponseWriter, r *http.Request) {
	if rt.handleCORS(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (rt *Runtime) Categories(w http.ResponseWriter, r *http.Request) {
	if rt.handleCORS(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	const query = `
SELECT
	id,
	COALESCE(name, '') AS name,
	COALESCE(short_name, '') AS short_name,
	now() AS created_at,
	now() AS updated_at
FROM public.category
ORDER BY id ASC
`

	rows, err := rt.db.QueryContext(r.Context(), query)
	if err != nil {
		writeInternalError(w)
		return
	}
	defer rows.Close()

	out := make([]Category, 0, 16)
	for rows.Next() {
		var item Category
		if err := rows.Scan(&item.ID, &item.Name, &item.ShortName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			writeInternalError(w)
			return
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		writeInternalError(w)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": out})
}

func (rt *Runtime) Subcategories(w http.ResponseWriter, r *http.Request) {
	if rt.handleCORS(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	categoryID, err := parsePositiveInt64(r.URL.Query().Get("category_id"), "category_id")
	if err != nil {
		writeValidationError(w, err.Error())
		return
	}

	const query = `
SELECT
	id,
	COALESCE(category_id, 0) AS category_id,
	COALESCE(name, '') AS name,
	now() AS created_at,
	now() AS updated_at
FROM public.subcategory
WHERE category_id = $1
ORDER BY id ASC
`

	rows, err := rt.db.QueryContext(r.Context(), query, categoryID)
	if err != nil {
		writeInternalError(w)
		return
	}
	defer rows.Close()

	out := make([]Subcategory, 0, 32)
	for rows.Next() {
		var item Subcategory
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.Name, &item.CreatedAt, &item.UpdatedAt); err != nil {
			writeInternalError(w)
			return
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		writeInternalError(w)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": out})
}

func (rt *Runtime) Posts(w http.ResponseWriter, r *http.Request) {
	if rt.handleCORS(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	filter, err := parsePostFilter(r)
	if err != nil {
		writeValidationError(w, err.Error())
		return
	}

	var (
		categoryID sql.NullInt64
		subcatID   sql.NullInt64
		status     sql.NullInt64
	)
	if filter.CategoryID > 0 {
		categoryID.Valid = true
		categoryID.Int64 = filter.CategoryID
	}
	if filter.SubcategoryID > 0 {
		subcatID.Valid = true
		subcatID.Int64 = filter.SubcategoryID
	}
	if filter.Status != 0 {
		status.Valid = true
		status.Int64 = int64(filter.Status)
	}

	const countQuery = `
SELECT COUNT(1)
FROM public.post
WHERE ($1::bigint IS NULL OR category_id = $1)
  AND ($2::bigint IS NULL OR subcategory_id = $2)
  AND ($3::int IS NULL OR status = $3)
`
	var total int
	if err := rt.db.QueryRowContext(r.Context(), countQuery, categoryID, subcatID, status).Scan(&total); err != nil {
		writeInternalError(w)
		return
	}

	const listQuery = `
SELECT
	id,
	COALESCE(category_id, 0) AS category_id,
	COALESCE(subcategory_id, 0) AS subcategory_id,
	COALESCE(email, '') AS email,
	COALESCE(name, '') AS name,
	COALESCE(body, '') AS body,
	COALESCE(status, 0) AS status,
	COALESCE(time_posted, 0) AS time_posted,
	COALESCE(time_posted_at, to_timestamp(0)) AS time_posted_at,
	COALESCE(price::float8, 0) AS price,
	(price IS NOT NULL) AS has_price,
	(
		COALESCE(photo1_file_name, '') <> '' OR
		COALESCE(photo2_file_name, '') <> '' OR
		COALESCE(photo3_file_name, '') <> '' OR
		COALESCE(photo4_file_name, '') <> '' OR
		COALESCE(image_source1, '') <> '' OR
		COALESCE(image_source2, '') <> '' OR
		COALESCE(image_source3, '') <> '' OR
		COALESCE(image_source4, '') <> ''
	) AS has_image,
	COALESCE(created_at, now()) AS created_at,
	COALESCE(updated_at, created_at, now()) AS updated_at
FROM public.post
WHERE ($1::bigint IS NULL OR category_id = $1)
  AND ($2::bigint IS NULL OR subcategory_id = $2)
  AND ($3::int IS NULL OR status = $3)
ORDER BY time_posted DESC NULLS LAST, id DESC
LIMIT $4 OFFSET $5
`

	rows, err := rt.db.QueryContext(r.Context(), listQuery, categoryID, subcatID, status, filter.Limit, filter.Offset)
	if err != nil {
		writeInternalError(w)
		return
	}
	defer rows.Close()

	out := make([]Post, 0, filter.Limit)
	for rows.Next() {
		var item Post
		if err := rows.Scan(
			&item.ID,
			&item.CategoryID,
			&item.SubcategoryID,
			&item.Email,
			&item.Name,
			&item.Body,
			&item.Status,
			&item.TimePosted,
			&item.TimePostedAt,
			&item.Price,
			&item.HasPrice,
			&item.HasImage,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			writeInternalError(w)
			return
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		writeInternalError(w)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": out,
		"meta": map[string]int{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

type postFilter struct {
	CategoryID    int64
	SubcategoryID int64
	Status        int
	Limit         int
	Offset        int
}

func parsePostFilter(r *http.Request) (postFilter, error) {
	query := r.URL.Query()
	filter := postFilter{
		Limit:  20,
		Offset: 0,
		Status: 1,
	}

	if query.Has("category_id") {
		v, err := parseNonNegativeInt64(query.Get("category_id"), "category_id")
		if err != nil {
			return filter, err
		}
		filter.CategoryID = v
	}
	if query.Has("subcategory_id") {
		v, err := parseNonNegativeInt64(query.Get("subcategory_id"), "subcategory_id")
		if err != nil {
			return filter, err
		}
		filter.SubcategoryID = v
	}
	if query.Has("status") {
		v, err := parseNonNegativeInt(query.Get("status"), "status")
		if err != nil {
			return filter, err
		}
		filter.Status = v
	}
	if query.Has("limit") {
		v, err := parsePositiveInt(query.Get("limit"), "limit")
		if err != nil {
			return filter, err
		}
		if v > 100 {
			v = 100
		}
		filter.Limit = v
	}
	if query.Has("offset") {
		v, err := parseNonNegativeInt(query.Get("offset"), "offset")
		if err != nil {
			return filter, err
		}
		filter.Offset = v
	}

	return filter, nil
}

func (rt *Runtime) handleCORS(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin != "" && isOriginAllowed(origin, rt.corsOrigins) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func buildAllowedOrigins(origins string) map[string]struct{} {
	values := map[string]struct{}{
		"http://localhost:3000": {},
		"http://127.0.0.1:3000": {},
	}
	for _, value := range strings.Split(origins, ",") {
		trimmed := strings.TrimSpace(strings.TrimSuffix(value, "/"))
		if trimmed == "" {
			continue
		}
		values[trimmed] = struct{}{}
	}
	return values
}

func isOriginAllowed(origin string, allowed map[string]struct{}) bool {
	origin = strings.TrimSuffix(origin, "/")
	_, ok := allowed[origin]
	return ok
}

func parsePositiveInt64(raw string, key string) (int64, error) {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return value, nil
}

func parseNonNegativeInt64(raw string, key string) (int64, error) {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", key)
	}
	return value, nil
}

func parsePositiveInt(raw string, key string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return value, nil
}

func parseNonNegativeInt(raw string, key string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", key)
	}
	return value, nil
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, errorEnvelope{
		Error: apiError{
			Code:    "method_not_allowed",
			Message: "method not allowed",
		},
	})
}

func writeValidationError(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, errorEnvelope{
		Error: apiError{
			Code:    "validation_error",
			Message: message,
		},
	})
}

func writeInternalError(w http.ResponseWriter) {
	writeJSON(w, http.StatusInternalServerError, errorEnvelope{
		Error: apiError{
			Code:    "internal_error",
			Message: "internal server error",
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func ensurePoolerSafeConnectionString(databaseURL string) string {
	trimmed := strings.TrimSpace(databaseURL)
	if trimmed == "" {
		return trimmed
	}

	if parsed, err := url.Parse(trimmed); err == nil && parsed.Scheme != "" {
		query := parsed.Query()
		if query.Get("default_query_exec_mode") == "" {
			query.Set("default_query_exec_mode", "simple_protocol")
		}
		if query.Get("statement_cache_capacity") == "" {
			query.Set("statement_cache_capacity", "0")
		}
		if query.Get("description_cache_capacity") == "" {
			query.Set("description_cache_capacity", "0")
		}
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}

	if !strings.Contains(trimmed, "default_query_exec_mode=") {
		trimmed += " default_query_exec_mode=simple_protocol"
	}
	if !strings.Contains(trimmed, "statement_cache_capacity=") {
		trimmed += " statement_cache_capacity=0"
	}
	if !strings.Contains(trimmed, "description_cache_capacity=") {
		trimmed += " description_cache_capacity=0"
	}
	return trimmed
}

func (rt *Runtime) Ping(ctx context.Context) error {
	return rt.db.PingContext(ctx)
}
