package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Capmus-Team/supost-cli-mark/internal/adapters"
	"github.com/Capmus-Team/supost-cli-mark/internal/config"
	"github.com/Capmus-Team/supost-cli-mark/internal/domain"
	"github.com/Capmus-Team/supost-cli-mark/internal/repository"
	"github.com/Capmus-Team/supost-cli-mark/internal/service"
	"github.com/spf13/cobra"
)

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a preview HTTP server",
	Long: `Start a lightweight HTTP server that exposes the service layer as JSON
endpoints. This is for prototyping only and powers frontend integration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		port := cfg.Port
		if port == 0 {
			port = 8080
		}

		var (
			marketRepo repository.MarketplaceStore
			closeRepo  func() error
		)
		if strings.TrimSpace(cfg.DatabaseURL) != "" {
			pgRepo, err := repository.NewPostgres(cfg.DatabaseURL)
			if err != nil {
				return fmt.Errorf("connecting to postgres: %w", err)
			}
			marketRepo = pgRepo
			closeRepo = pgRepo.Close
		} else {
			memRepo := repository.NewInMemory()
			marketRepo = memRepo
		}
		if closeRepo != nil {
			defer func() {
				_ = closeRepo()
			}()
		}

		var emailSender adapters.EmailSender
		if cfg.MailgunAPIKey != "" && cfg.MailgunDomain != "" && cfg.MailgunFrom != "" {
			emailSender = adapters.NewMailgunSender(cfg.MailgunAPIKey, cfg.MailgunDomain, cfg.MailgunFrom)
		}
		marketSvc := service.NewMarketplaceService(marketRepo, emailSender)

		mux := http.NewServeMux()
		registerHandlers(mux, marketSvc)

		handler := withCORS(cfg.CORSOrigins, mux)
		addr := fmt.Sprintf(":%d", port)

		logger := slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), nil))
		logger.Info("preview server started",
			"addr", "http://localhost"+addr,
			"routes", "/api/health,/api/categories,/api/subcategories,/api/posts,/api/posts/{id},/api/posts/{id}/messages",
		)
		return http.ListenAndServe(addr, handler)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntP("port", "p", 8080, "port to listen on")
}

func registerHandlers(mux *http.ServeMux, marketSvc *service.MarketplaceService) {
	mux.HandleFunc("GET /api/categories", func(w http.ResponseWriter, r *http.Request) {
		categories, err := marketSvc.ListCategories(r.Context())
		if err != nil {
			writeInternalError(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": categories})
	})

	mux.HandleFunc("GET /api/subcategories", func(w http.ResponseWriter, r *http.Request) {
		categoryID, err := parsePositiveInt64(r.URL.Query().Get("category_id"), "category_id")
		if err != nil {
			writeValidationError(w, err.Error())
			return
		}

		subcategories, err := marketSvc.ListSubcategories(r.Context(), categoryID)
		if err != nil {
			if errors.Is(err, domain.ErrValidation) {
				writeValidationError(w, err.Error())
				return
			}
			writeInternalError(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": subcategories})
	})

	mux.HandleFunc("GET /api/posts", func(w http.ResponseWriter, r *http.Request) {
		filter, err := parsePostFilter(r)
		if err != nil {
			writeValidationError(w, err.Error())
			return
		}

		posts, total, err := marketSvc.ListPosts(r.Context(), filter)
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
	})

	mux.HandleFunc("GET /api/posts/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			writeValidationError(w, "post id must be a positive integer")
			return
		}

		post, err := marketSvc.GetPost(r.Context(), id)
		if err != nil {
			if errors.Is(err, domain.ErrValidation) {
				writeValidationError(w, err.Error())
				return
			}
			writeInternalError(w)
			return
		}
		if post == nil {
			writeJSON(w, http.StatusNotFound, errorEnvelope{
				Error: apiError{
					Code:    "not_found",
					Message: "post not found",
				},
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{"data": post})
	})

	mux.HandleFunc("POST /api/posts/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		postID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || postID <= 0 {
			writeValidationError(w, "post id must be a positive integer")
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			writeValidationError(w, "Content-Type must be application/json")
			return
		}

		var body struct {
			Message string `json:"message"`
			Email   string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeValidationError(w, "invalid JSON body")
			return
		}

		email := strings.TrimSpace(body.Email)
		message := strings.TrimSpace(body.Message)
		if email == "" {
			writeValidationError(w, "email is required")
			return
		}
		if message == "" {
			writeValidationError(w, "message is required")
			return
		}
		if len(email) > 70 {
			writeValidationError(w, "email too long")
			return
		}
		if len(message) > 32000 {
			writeValidationError(w, "message too long")
			return
		}

		// Verify post exists
		post, err := marketSvc.GetPost(r.Context(), postID)
		if err != nil || post == nil {
			writeJSON(w, http.StatusNotFound, errorEnvelope{
				Error: apiError{Code: "not_found", Message: "post not found"},
			})
			return
		}

		if err := marketSvc.SendMessage(r.Context(), postID, email, message); err != nil {
			logger := slog.Default()
			logger.Error("send message failed", "err", err, "post_id", postID)
			writeJSON(w, http.StatusInternalServerError, errorEnvelope{
				Error: apiError{
					Code:    "send_failed",
					Message: err.Error(),
				},
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{"data": map[string]string{"status": "sent"}})
	})

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
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

func withCORS(origins string, next http.Handler) http.Handler {
	allowedOrigins := buildAllowedOrigins(origins)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isOriginAllowed(origin, allowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
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
	if _, ok := allowed[origin]; ok {
		return true
	}
	return false
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
