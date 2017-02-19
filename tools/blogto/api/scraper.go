package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gosimple/slug"
	"github.com/goware/lg"
)

const (
	API_URL = "http://api.blogto.com/listings.json"
	KEY     = "d4UKy3LrxPXA"
)

func Worker(num int, pageCh, resCh chan int) {
	for i := range pageCh {
		lg.Infof("worker %d scraping page %d", num, i)

		res, err := http.Get(fmt.Sprintf("%s?k=%s&page=%d", API_URL, KEY, i))
		if err != nil {
			lg.Warnf("page(%d) errored: %+v", i, err)
		}

		storeFile, err := os.Create(fmt.Sprintf("data/%d.json", i))
		if err != nil {
			lg.Warnf("errored create file(%d)", i)
		}

		io.Copy(storeFile, res.Body)

		storeFile.Close()
		res.Body.Close()

		resCh <- 1
	}
}

func Scraper() {
	pageCh := make(chan int, 400)
	resCh := make(chan int, 400)

	for i := 0; i < 10; i++ {
		go Worker(i+1, pageCh, resCh)
	}

	for i := 1; i <= 400; i++ {
		pageCh <- i
	}
	close(pageCh)

	for i := 1; i <= 400; i++ {
		<-resCh
	}
}

func main() {
	type wrapper struct {
		Results []*struct {
			Name    string `json:"name"`
			Excerpt string `json:"excerpt"`
		} `json:"results"`
	}

	var excerpts []map[string]string
	for i := 1; i <= 400; i++ {
		f, err := os.Open(fmt.Sprintf("data/%d.json", i))
		if err != nil {
			lg.Warn(err)
			continue
		}

		var res *wrapper
		if err := json.NewDecoder(f).Decode(&res); err != nil {
			lg.Warnf("page(%d) error: %v", i, err)
			continue
		}

		for _, s := range res.Results {
			excerpts = append(excerpts, map[string]string{slug.Make(s.Name): s.Excerpt})
		}

		f.Close()
	}

	x, _ := os.Create("out.json")
	json.NewEncoder(x).Encode(&excerpts)

	x.Close()
}
