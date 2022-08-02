package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/StalkR/imdb"
)

func ImdbSearch(w http.ResponseWriter, r *http.Request) {
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
		result, err := SearchByTitle(title)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(toError(err.Error()))
			return
		}
		if len(result) == 0 {
			w.WriteHeader(http.StatusNotFound)
			w.Write(toError("Not found"))
		} else {
			w.Write([]byte(toJson(result)))
		}
	} else {
		result, err := imdb.NewTitle(http.DefaultClient, id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write(toError(err.Error()))
		} else {
			w.Write([]byte(toJson(result)))
		}
	}
}

func toJson(data interface{}) string {
	b, _ := json.Marshal(data)
	return string(b)
}

func toError(err string) []byte {
	response := map[string]string{
		"error": err,
	}
	return []byte(toJson(response))
}

type Title struct {
	Title  string `json:"title,omitempty"`
	ID     string `json:"id,omitempty"`
	Year   string `json:"year,omitempty"`
	Actors string `json:"actors,omitempty"`
	Rank   string `json:"rank,omitempty"`
	Link   string `json:"link,omitempty"`
	Poster string `json:"poster,omitempty"`
}

type rawData struct {
	D []struct {
		I struct {
			ImageURL string `json:"imageUrl"`
		} `json:"i,omitempty"`
		ID   string `json:"id"`
		L    string `json:"l"`
		Q    string `json:"q"`
		Rank int    `json:"rank"`
		S    string `json:"s"`
		Vt   int    `json:"vt,omitempty"`
		Y    int    `json:"y"`
	} `json:"d"`
}

func SearchByTitle(q string) ([]Title, error) {
	firstLetter := strings.ToLower(string(q[0]))
	URL := "https://v2.sg.media-imdb.com/suggestion/titles/" + firstLetter + "/" + url.QueryEscape(q) + ".json"
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data rawData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	var titles []Title
	for _, r := range data.D {
		titles = append(titles, Title{Title: r.L, Year: fmt.Sprint(r.Y), ID: r.ID, Actors: r.S, Rank: fmt.Sprint(r.Rank), Link: "https://www.imdb.com/title/" + r.ID, Poster: r.I.ImageURL})
	}
	return titles, nil
}
