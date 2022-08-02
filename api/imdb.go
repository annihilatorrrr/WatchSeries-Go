package main

import (
	"encoding/json"
	"net/http"

	"github.com/StalkR/imdb"
)

func ImdbTitle(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	id := r.URL.Query().Get("id")
	query := r.URL.Query().Get("query")
	if title == "" && id == "" && query == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if title == "" && query != "" {
		title = query
	}
	if title != "" {
		result, _ := imdb.SearchTitle(http.DefaultClient, title)
		if len(result) == 0 {
			response := map[string]string{
				"error": "No results found",
			}
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(toJson(response)))
		} else {
			w.Write([]byte(toJson(result)))
		}
	} else {
		result, err := imdb.NewTitle(http.DefaultClient, id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			response := map[string]string{
				"error": err.Error(),
			}
			w.Write([]byte(toJson(response)))
		} else {
			w.Write([]byte(toJson(result)))
		}
	}
}

func toJson(data interface{}) string {
	b, _ := json.Marshal(data)
	return string(b)
}
