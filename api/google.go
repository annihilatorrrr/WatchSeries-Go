package handler

import (
	"context"
	"encoding/json"
	"net/http"

	gs "github.com/rocketlaunchr/google-search"
)

func Google(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	q := query.Get("q")
	i := query.Get("i")
	if query.Get("q") == "" {
		http.Error(w, "missing query", http.StatusBadRequest)
		return
	}
	search, err := gs.Search(context.TODO(), q, gs.SearchOptions{Limit: 10})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if i == "true" {
		b, err := json.MarshalIndent(search, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(b)
		return
	}
	json.NewEncoder(w).Encode(search)
}
