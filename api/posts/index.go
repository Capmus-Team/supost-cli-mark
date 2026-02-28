package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/Capmus-Team/supost-cli/internal/vercelapi"
)

var (
	apiInstance *vercelapi.API
	apiErr      error
	apiOnce     sync.Once
)

func main() {}

func Handler(w http.ResponseWriter, r *http.Request) {
	apiOnce.Do(func() {
		apiInstance, apiErr = vercelapi.NewAPI()
	})
	if apiErr != nil {
		log.Printf("initializing api: %v", apiErr)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	apiInstance.Posts(w, r)
}
