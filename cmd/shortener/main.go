package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

func postHandler(u url) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodPost:
			body, err := c.GetRawData()

			if err != nil || len(body) == 0 {
				c.String(http.StatusBadRequest, "Invalid Body")
				return
			}
			url := string(body)
			randomStr := getRandString(8)
			u.postURL(randomStr, url)
			c.Header("Content-Type", "text/plain")
			c.String(http.StatusCreated, "http://localhost:8080/%s", randomStr)

		case http.MethodGet:
			shortURL := c.Param("short")
			original, exists := u.getURL(shortURL)
			if !exists {
				c.String(http.StatusBadRequest, "Not found Url")
			}
			c.Header("Content-Type", "text/plain")
			c.Redirect(http.StatusTemporaryRedirect, original)

		default:
			c.String(http.StatusMethodNotAllowed, "Invalid Method")
		}

	}
}
func main() {
	r := gin.Default()
	storage := NewURLStorage()
	r.POST("/", postHandler(storage))
	r.GET("/:short", postHandler(storage))
	fmt.Println("Listening port 8080....")
	r.Run(":8080")

}
