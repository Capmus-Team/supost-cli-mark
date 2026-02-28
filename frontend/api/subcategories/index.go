package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/Capmus-Team/supost-cli-mark/frontend/apiruntime"
)

var (
	apiInstance *apiruntime.Runtime
	apiErr      error
	apiOnce     sync.Once
)

func Handler(w http.ResponseWriter, r *http.Request) {
	apiOnce.Do(func() {
		apiInstance, apiErr = apiruntime.GetRuntime()
	})
	if apiErr != nil {
		log.Printf("initializing api: %v", apiErr)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	apiInstance.Subcategories(w, r)
}
