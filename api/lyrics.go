package handler

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

func LyricsHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		q = r.URL.Query().Get("query")
	}
	if q == "" {
		http.Error(w, "no query", http.StatusBadRequest)
		return
	}
	lyric, err := searchLyrics(q)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if r.URL.Query().Get("lyrics") == "true" {
		var l string
		for _, line := range lyric.Lyrics.Lines {
			l += line.Words + " "
		}
		w.Write([]byte(l))
	} else {
		json.NewEncoder(w).Encode(lyric)
	}
}

type (
	spotifyClient struct {
		AccessToken string `json:"access_token"`
		Expire      string `json:"expire"`
		HttpCli     *http.Client
	}
)

type SpotifyResult struct {
	Data struct {
		SearchV2 struct {
			Tracks struct {
				TotalCount int `json:"totalCount"`
				Items      []struct {
					Data struct {
						Typename     string `json:"__typename"`
						URI          string `json:"uri"`
						ID           string `json:"id"`
						Name         string `json:"name"`
						AlbumOfTrack struct {
							URI      string `json:"uri"`
							Name     string `json:"name"`
							CoverArt struct {
								Sources []struct {
									URL    string `json:"url"`
									Width  int    `json:"width"`
									Height int    `json:"height"`
								} `json:"sources"`
							} `json:"coverArt"`
							ID string `json:"id"`
						} `json:"albumOfTrack"`
						Artists struct {
							Items []struct {
								URI     string `json:"uri"`
								Profile struct {
									Name string `json:"name"`
								} `json:"profile"`
							} `json:"items"`
						} `json:"artists"`
						ContentRating struct {
							Label string `json:"label"`
						} `json:"contentRating"`
						Duration struct {
							TotalMilliseconds int `json:"totalMilliseconds"`
						} `json:"duration"`
						Playability struct {
							Playable bool `json:"playable"`
						} `json:"playability"`
					} `json:"data"`
				} `json:"items"`
			} `json:"tracks"`
		} `json:"searchV2"`
	} `json:"data"`
}

type Lyric struct {
	Lyrics struct {
		SyncType string `json:"syncType"`
		Lines    []struct {
			StartTimeMs string        `json:"startTimeMs"`
			Words       string        `json:"words"`
			Syllables   []interface{} `json:"syllables"`
			EndTimeMs   string        `json:"endTimeMs"`
		} `json:"lines"`
		Provider            string        `json:"provider"`
		ProviderLyricsID    string        `json:"providerLyricsId"`
		ProviderDisplayName string        `json:"providerDisplayName"`
		SyncLyricsURI       string        `json:"syncLyricsUri"`
		IsDenseTypeface     bool          `json:"isDenseTypeface"`
		Alternatives        []interface{} `json:"alternatives"`
		Language            string        `json:"language"`
		IsRtlLanguage       bool          `json:"isRtlLanguage"`
		FullscreenAction    string        `json:"fullscreenAction"`
	} `json:"lyrics"`
	Colors struct {
		Background    int `json:"background"`
		Text          int `json:"text"`
		HighlightText int `json:"highlightText"`
	} `json:"colors"`
	HasVocalRemoval bool `json:"hasVocalRemoval"`
}

func (sp *spotifyClient) init() error {
	sp.HttpCli = &http.Client{Timeout: 10 * time.Second}
	err := sp.setSpotifyCred()
	if err != nil {
		return err
	}
	return nil
}

func (sp *spotifyClient) setCookies(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
	req.Header.Set("cookie", "sp_m=in-en; sp_t=68e5d30e-56c8-4e76-859b-f93b9aa5b64e; sp_ab=%7B%222019_04_premium_menu%22%3A%22control%22%7D; spot=%7B%22t%22%3A1663244415%2C%22m%22%3A%22in-en%22%2C%22p%22%3Anull%7D; sp_dc=AQA49YQYf_994zc5sKlhYUM2ZCAnHPNe6eZ7h5e_G95YyZzTyu_1WRvyaKQDe2JBWmuYzkBfgI5epmQASfnPx4ju4pgt7EGlBsBwCN8NHi030Khf0CdtBV8LspofUMd-deDZuT5wOOojfREYFxL0Zr_8eG-Vu_z-; OptanonAlertBoxClosed=2022-09-17T10:22:53.741Z; sp_landing=https%3A%2F%2Fwww.spotify.com%2Fin-en%2F; mwpA2HS_status=complete; OptanonConsent=isIABGlobal=false&datestamp=Wed+Sep+28+2022+12%3A52%3A40+GMT%2B0530+(India+Standard+Time)&version=6.26.0&hosts=&landingPath=NotLandingPage&groups=s00%3A1%2Cf00%3A1%2Cm00%3A1%2Ct00%3A1%2Ci00%3A1%2Cf02%3A1%2Cm02%3A1%2Ct02%3A1&AwaitingReconsent=false&geolocation=IN%3BKL")
}

func (sp *spotifyClient) setSpotifyCred() error {
	if sp.AccessToken == "" {
		token, expire, err := sp.genSpotifyToken()
		if err != nil {
			return err
		}
		sp.AccessToken = token
		sp.Expire = expire
		return nil
	} else {
		exp, _ := strconv.ParseInt(sp.Expire, 10, 64)
		if time.Now().Unix() > (exp / 1000) {
			token, expire, err := sp.genSpotifyToken()
			if err != nil {
				return err
			}
			sp.AccessToken = token
			sp.Expire = expire
			return nil
		} else {
			return nil
		}
	}
}

func (sp *spotifyClient) genSpotifyToken() (string, string, error) {
	req, _ := http.NewRequest("GET", "https://open.spotify.com/", nil)
	sp.setCookies(req)
	resp, err := sp.HttpCli.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	accessTokenReg := regexp.MustCompile(`"accessToken":"(.+?)"`)
	expireTimeReg := regexp.MustCompile(`"accessTokenExpirationTimestampMs":(\d+)`)
	var body []byte
	body, _ = ioutil.ReadAll(resp.Body)
	accessToken := accessTokenReg.FindStringSubmatch(string(body))
	expireTime := expireTimeReg.FindStringSubmatch(string(body))
	if len(accessToken) > 1 {
		return accessToken[1], expireTime[1], nil
	} else {
		return "", "", nil
	}
}

func (sp *spotifyClient) searchSpotify(query string) (SpotifyResult, error) {
	req, _ := http.NewRequest("GET", `https://api-partner.spotify.com/pathfinder/v1/query?operationName=searchDesktop&variables=%7B%22searchTerm%22%3A%22`+url.QueryEscape(query)+`%22%2C%22offset%22%3A0%2C%22limit%22%3A10%2C%22numberOfTopResults%22%3A5%2C%22includeAudiobooks%22%3Afalse%7D&extensions=%7B%22persistedQuery%22%3A%7B%22version%22%3A1%2C%22sha256Hash%22%3A%2219967195df75ab8b51161b5ac4586eab9cf73b51b35a03010073533c33fd11ae%22%7D%7D`, nil)
	req.Header.Set("app-platform", "WebPlayer")
	req.Header.Set("authorization", "Bearer "+sp.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SpotifyResult{}, err
	}
	defer resp.Body.Close()
	var body SpotifyResult
	json.NewDecoder(resp.Body).Decode(&body)
	return body, nil
}

func (sp *spotifyClient) lyrics(trackID string) (*Lyric, error) {
	req, _ := http.NewRequest("GET", "https://spclient.wg.spotify.com/color-lyrics/v2/track/"+trackID+"/image/https%3A%2F%2Fi.scdn.co%2Fimage%2Fab67616d0000b2738dce351c5e4a62c2ea2dd498?format=json&vocalRemoval=false&market=from_token", nil)
	req.Header.Set("authorization", "Bearer "+sp.AccessToken)
	req.Header.Set("app-platform", "WebPlayer")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36")
	req.Header.Set("client-token", "AABExvHOScjFTTqs3FS13iTJVAiiEgQVQ0n6U3Q8+mzW8GkrZqfQ5imXsbMxHiHEIlaRI5oEkHU4lwvxqmSw65pg5gwMBPzQkb80HtNExDRxPYDPb2k5lPogRv6Kmir9LE4m/GSe+nq4vEV+ADNSHta1IKN0kIjUz50115LRfaTf2+1spEZn5DOnGFcBXHj9Gh4KcTzm8OVsBNJ8lGgwHhNjTko0CjxjEbLRRGUnRGemZ5a0aAtfEacfq9t4/+d9DEbPo2AcZ3CVL6dI80MkYNi2jvxP0zDQnbUBKlfVQdI=")
	resp, err := sp.HttpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var body Lyric
	json.NewDecoder(resp.Body).Decode(&body)
	return &body, nil
}

func searchLyrics(q string) (*Lyric, error) {
	sptfy := spotifyClient{}
	if err := sptfy.init(); err != nil {
		return nil, err
	}
	res, err := sptfy.searchSpotify(q)
	if err != nil {
		return nil, err
	}
	if res.Data.SearchV2.Tracks.TotalCount == 0 {
		return nil, errors.New("no result")
	}
	lyric, err := sptfy.lyrics(res.Data.SearchV2.Tracks.Items[0].Data.ID)
	if err != nil {
		return nil, err
	}
	return lyric, nil
}
