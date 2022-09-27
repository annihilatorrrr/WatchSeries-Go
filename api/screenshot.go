package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func ScreenShot(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	bytesOfImage, err := Screenshot(url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(toError(err.Error()))
		return
	}
	w.Write(bytesOfImage)
	w.Header().Set("Content-Type", "image/png")
}

func Screenshot(uri string) ([]byte, error) {
	apiURL := "https://webshot.deam.io/%s"
	URI, _ := url.Parse(fmt.Sprintf(apiURL, uri))
	params := url.Values{}
	params.Add("type", "jpeg")
	params.Add("width", "1280")
	params.Add("height", "720")
	params.Add("fullPage", "true")
	params.Add("quality", "100")
	params.Add("delay", "0")
	URI.RawQuery = params.Encode()
	resp, err := http.Get(URI.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bytesOfImage, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytesOfImage, nil
}
