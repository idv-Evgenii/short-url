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
	postUrl(short, original string)
	getUrl(short string) (string, bool)
}
type URLStorage struct {
	urlmap map[string]string
}

func NewUrlStorage() *URLStorage {
	return &URLStorage{
		urlmap: make(map[string]string),
	}
}

func (u *URLStorage) postUrl(short, original string) {
	u.urlmap[short] = original
}

func (u *URLStorage) getUrl(short string) (string, bool) {
	original, exists := u.urlmap[short]
	return original, exists
}

func getRandString(n int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, n)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
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
			u.postUrl(randomStr, string(urlName))
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(fmt.Sprintf("Your new url: %s\r\n", randomStr)))

			// http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
			// return
		} else if r.Method == http.MethodGet {
			path := r.URL.Path[1:]
			originalUrl, exists := u.getUrl(path)
			if !exists {
				http.Error(w, "Not found Url ", 400)
			} else {
				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("Location", originalUrl)
				w.WriteHeader(http.StatusTemporaryRedirect)
				w.Write([]byte(fmt.Sprintf("Original URL: %s\r\n", originalUrl)))
			}

		}

	}
}

func main() {
	storage := NewUrlStorage()
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, postHandler(storage))
	fmt.Println("Listening port 8080....")
	log.Fatal(http.ListenAndServe(`:8080`, mux))
}
