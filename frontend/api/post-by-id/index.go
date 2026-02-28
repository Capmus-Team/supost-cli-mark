package handler

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/Capmus-Team/supost-cli-mark/frontend/apiruntime"
)

var (
	postAPIInstance *apiruntime.Runtime
	postAPIErr      error
	postAPIOnce     sync.Once
)

func Handler(w http.ResponseWriter, r *http.Request) {
	postAPIOnce.Do(func() {
		postAPIInstance, postAPIErr = apiruntime.GetRuntime()
	})
	if postAPIErr != nil {
		log.Printf("initializing api: %v", postAPIErr)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "post id required", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	postAPIInstance.GetPost(w, r, id)
}
