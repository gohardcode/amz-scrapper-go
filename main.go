package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/julienschmidt/httprouter"
)

type ProductInfo struct {
	Title   string `json:"title"`
	Price   string `json:"price"`
	Image   string `json:"image"`
	InStock bool   `json:"in_stock"`
}

func ParseProductPage(r io.Reader) (*ProductInfo, error) {
	doc, err := goquery.NewDocumentFromReader(r)

	if err != nil {
		return nil, err
	}

	return &ProductInfo{
		Title: strings.TrimSpace(doc.Find("#productTitle").Text()),
		Price: strings.TrimSpace(doc.Find("#buyNewSection .offer-price").Text()),
		Image: func() string {
			data, ok := doc.Find("#leftCol img[data-a-dynamic-image]").First().Attr("data-a-dynamic-image")
			if !ok {
				return ""
			}

			// <img> attribute "data-a-dynamic-image" contains
			// json object of type { "url":[width,height], "..."}
			images := map[string][2]int{}

			if err := json.Unmarshal([]byte(data), &images); err != nil || len(images) == 0 {
				return ""
			}

			// so i try to get image with the max width
			maxWidth := 0
			res := ""

			for src, size := range images {
				if size[0] < maxWidth {
					continue
				}

				maxWidth = size[0]
				res = src
			}

			return res
		}(),
		InStock: strings.TrimSpace(doc.Find("#availability").Text()) == "In stock.",
	}, nil

}

type Product struct {
	URL   string       `json:"url"`
	Meta  *ProductInfo `json:"meta,omitempty"`
	Error string       `json:"error,omitempty"`
}

func HandlePayload(payload []string) []Product {

	res := make([]Product, len(payload))

	for i, URL := range payload {

		res[i].URL = URL

		p, err := func() (*ProductInfo, error) {

			if _, err := url.ParseRequestURI(URL); err != nil {
				return nil, err
			}

			resp, err := http.Get(URL)

			if err == nil && resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("%s", http.StatusText(resp.StatusCode))
			}

			if err != nil {
				return nil, err
			}

			return ParseProductPage(resp.Body)
		}()

		if err != nil {
			res[i].Error = err.Error()
			continue
		}

		res[i].Meta = p
	}

	return res
}

type proc struct {
	result []Product
	done   chan struct{}
}

var ps = map[string]*proc{}
var psrwmu = sync.RWMutex{}

func main() {

	router := httprouter.New()

	router.POST("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		d := json.NewDecoder(r.Body)
		var payload []string

		if err := d.Decode(&payload); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		res := HandlePayload(payload)
		out, _ := json.Marshal(res)

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	})

	router.POST("/:requestID", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

		d := json.NewDecoder(r.Body)
		var payload []string

		if err := d.Decode(&payload); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		reqID := params.ByName("requestID")

		psrwmu.Lock()
		defer psrwmu.Unlock()

		if _, exists := ps[reqID]; exists {
			http.Error(w, "cannot overwrite", http.StatusConflict)
			return
		}

		p := &proc{
			done: make(chan struct{}),
		}
		ps[reqID] = p

		go func() {
			log.Printf("%s: started", reqID)
			res := HandlePayload(payload)
			p.result = res
			close(p.done)
			log.Printf("%s: finished", reqID)
		}()

		w.WriteHeader(http.StatusCreated)
		return
	})

	router.GET("/:requestID", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

		reqID := params.ByName("requestID")

		psrwmu.RLock()
		p, ok := ps[reqID]
		psrwmu.RUnlock()

		if !ok {
			http.NotFound(w, r)
			return
		}

		<-p.done
		psrwmu.Lock()
		delete(ps, reqID)
		psrwmu.Unlock()

		out, _ := json.Marshal(p.result)

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	})

	log.Println("listen on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
