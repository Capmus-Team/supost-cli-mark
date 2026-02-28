package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/Capmus-Team/supost-cli/frontend/api/shared"
)

var (
	apiInstance *shared.Runtime
	apiErr      error
	apiOnce     sync.Once
)

func Handler(w http.ResponseWriter, r *http.Request) {
	apiOnce.Do(func() {
		apiInstance, apiErr = shared.GetRuntime()
	})
	if apiErr != nil {
		log.Printf("initializing api: %v", apiErr)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	apiInstance.Categories(w, r)
}
