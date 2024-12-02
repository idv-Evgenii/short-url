package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type url interface {
	postURL(short, original string)
	getURL(short string) (string, bool)
}
type URLStorage struct {
	urlmap map[string]string
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		urlmap: make(map[string]string),
	}
}

func (u *URLStorage) postURL(short, original string) {
	u.urlmap[short] = original
}

func (u *URLStorage) getURL(short string) (string, bool) {
	original, exists := u.urlmap[short]
	return original, exists
}

func getRandString(n int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, n)
	for i := range result {
		result[i] = chars[rng.Intn(len(chars))]
	}
	return string(result)
}

func postHandler(u url) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "text/plain")
			urlName, err := io.ReadAll(r.Body)
			if err != nil || len(urlName) == 0 {
				http.Error(w, "Empty body", http.StatusBadRequest)
			}
			var randomStr = getRandString(8)
			u.postURL(randomStr, string(urlName))
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(fmt.Sprintf("http://localhost:8080/%s", randomStr)))

			// http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
			// return
		} else if r.Method == http.MethodGet {
			path := r.URL.Path[1:]
			originalURL, exists := u.getURL(path)
			if !exists {
				http.Error(w, "Not found Url ", http.StatusBadRequest)
			} else {
				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("Location", originalURL)
				w.WriteHeader(http.StatusTemporaryRedirect)
				w.Write([]byte(originalURL))
			}

		}

	}
}

func main() {
	storage := NewURLStorage()
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, postHandler(storage))
	fmt.Println("Listening port 8080....")
	log.Fatal(http.ListenAndServe(`:8080`, mux))
}
