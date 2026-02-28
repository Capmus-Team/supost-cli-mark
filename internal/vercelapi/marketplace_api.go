package vercelapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Capmus-Team/supost-cli/internal/config"
	"github.com/Capmus-Team/supost-cli/internal/domain"
	"github.com/Capmus-Team/supost-cli/internal/repository"
	"github.com/Capmus-Team/supost-cli/internal/service"
)

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type API struct {
	service *service.MarketplaceService
	cfg     *config.Config
}

func NewAPI() (*API, error) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	var repo repository.MarketplaceStore
	if strings.TrimSpace(cfg.DatabaseURL) != "" {
		pgRepo, err := repository.NewPostgres(cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("connecting to postgres: %w", err)
		}
		repo = pgRepo
	} else {
		repo = repository.NewInMemory()
	}

	return &API{
		service: service.NewMarketplaceService(repo),
		cfg:     cfg,
	}, nil
}

func (a *API) Categories(w http.ResponseWriter, r *http.Request) {
	if handleCORS(a.cfg.CORSOrigins, w, r) {
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	categories, err := a.service.ListCategories(r.Context())
	if err != nil {
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": categories})
}

func (a *API) Subcategories(w http.ResponseWriter, r *http.Request) {
	if handleCORS(a.cfg.CORSOrigins, w, r) {
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

	subcategories, err := a.service.ListSubcategories(r.Context(), categoryID)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			writeValidationError(w, err.Error())
			return
		}
		writeInternalError(w)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": subcategories})
}

func (a *API) Posts(w http.ResponseWriter, r *http.Request) {
	if handleCORS(a.cfg.CORSOrigins, w, r) {
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

	posts, total, err := a.service.ListPosts(r.Context(), filter)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			writeValidationError(w, err.Error())
			return
		}
		writeInternalError(w)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": posts,
		"meta": map[string]int{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

func (a *API) Health(w http.ResponseWriter, r *http.Request) {
	if handleCORS(a.cfg.CORSOrigins, w, r) {
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func parsePostFilter(r *http.Request) (domain.PostFilter, error) {
	query := r.URL.Query()
	filter := domain.PostFilter{
		Limit:  20,
		Offset: 0,
		Status: domain.PostStatusActive,
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

func handleCORS(origins string, w http.ResponseWriter, r *http.Request) bool {
	allowedOrigins := buildAllowedOrigins(origins)

	origin := r.Header.Get("Origin")
	if origin != "" && isOriginAllowed(origin, allowedOrigins) {
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
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		values[trimmed] = struct{}{}
	}
	return values
}

func isOriginAllowed(origin string, allowed map[string]struct{}) bool {
	_, ok := allowed[origin]
	return ok
}
