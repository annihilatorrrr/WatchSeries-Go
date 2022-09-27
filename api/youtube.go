package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/buger/jsonparser"
)

func Youtube(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	q := query.Get("q")
	if q == "" {
		q = query.Get("query")
	}
	i := query.Get("i")
	if query.Get("q") == "" {
		http.Error(w, "missing query", http.StatusBadRequest)
		return
	}
	search, err := YtSearch(q)
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
	w.Write(search)
}

func YtSearch(q string) ([]byte, error) {
	URL := "https://www.youtube.com/results?search_query=" + url.QueryEscape(q)
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	var exp, _ = regexp.Compile(`ytInitialData = [\s\S]*]`)
	b, _ := ioutil.ReadAll(resp.Body)
	match := exp.FindStringSubmatch(string(b))
	var d string
	if len(match) != 0 {
		d = match[0]
		d = strings.Replace(d, "ytInitialData = ", "", 1)
		d = strings.Split(d, ";</script>")[0]
	}
	pData := ParseYoutubeRAW(d)
	return pData, nil
}

func ParseYoutubeRAW(raw string) []byte {
	by := []byte(raw)
	var Results []YoutubeResult
	a, _, _, _ := jsonparser.Get(by, "contents", "twoColumnSearchResultsRenderer", "primaryContents", "sectionListRenderer", "contents")
	jsonparser.ArrayEach(a, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		b, _, _, _ := jsonparser.Get(value, "itemSectionRenderer", "contents")
		jsonparser.ArrayEach(b, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			c, _, _, _ := jsonparser.Get(value, "videoRenderer")
			d, _, _, _ := jsonparser.Get(c, "title", "runs")
			if d != nil {
				var Result YoutubeResult
				jsonparser.ArrayEach(d, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
					text, _, _, _ := jsonparser.Get(value, "text")
					Result.Title = string(text)
				})
				e, _, _, _ := jsonparser.Get(c, "thumbnail", "thumbnails", "[0]", "url")
				if e != nil {
					Result.Thumbnail = string(e)
				}
				f, _, _, _ := jsonparser.Get(c, "videoId")
				if f != nil {
					Result.URL = "https://www.youtube.com/watch?v=" + string(f)
				}
				metadata, _, _, _ := jsonparser.Get(c, "detailedMetadataSnippets")
				if metadata != nil {
					jsonparser.ArrayEach(metadata, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
						g, _, _, _ := jsonparser.Get(value, "snippetText", "runs")
						if g != nil {
							var desc string
							jsonparser.ArrayEach(g, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
								text, _, _, _ := jsonparser.Get(value, "text")
								desc += string(text)
							})
							Result.Description = desc
						}
					})
				}
				ownerText, _, _, _ := jsonparser.Get(c, "ownerText", "runs")
				if ownerText != nil {
					jsonparser.ArrayEach(ownerText, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
						text, _, _, _ := jsonparser.Get(value, "text")
						Result.Channel = string(text)
					})
				}
				videoID, _, _, _ := jsonparser.Get(c, "videoId")
				if videoID != nil {
					Result.ID = string(videoID)
				}
				published, _, _, _ := jsonparser.Get(c, "publishedTimeText", "simpleText")
				if published != nil {
					Result.Published = string(published)
				}
				length, _, _, _ := jsonparser.Get(c, "lengthText", "simpleText")
				if length != nil {
					Result.Duration = string(length)
				}
				views, _, _, _ := jsonparser.Get(c, "viewCountText", "simpleText")
				if views != nil {
					Result.Views = string(views)
				}
				Results = append(Results, Result)
			}
		})
	})
	data, _ := json.Marshal(Results)
	return data
}

type YoutubeResult struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
	Channel     string `json:"channel,omitempty"`
	Published   string `json:"published,omitempty"`
	Duration    string `json:"duration,omitempty"`
	Thumbnail   string `json:"thumbnail,omitempty"`
	Views       string `json:"views,omitempty"`
}
