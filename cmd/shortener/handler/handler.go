package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/idv-Evgenii/short-url/cmd/shortener/storage"
)

func PostHandler(u *storage.URLStorage, baseURL string) gin.HandlerFunc {
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
			_ = u.PostURL(randomStr, url) // Теперь uuid используется в методе postURL, но мы не выводим его напрямую
			c.Header("Content-Type", "text/plain")
			c.String(http.StatusCreated, "%s/%s", baseURL, randomStr)
		case http.MethodGet:
			shortURL := c.Param("short")
			original, exists := u.GetURL(shortURL)
			if !exists {
				c.String(http.StatusBadRequest, "Not found Url")
				return
			}
			c.Header("Content-Type", "text/plain")
			c.Redirect(http.StatusTemporaryRedirect, original)
		default:
			c.String(http.StatusMethodNotAllowed, "Invalid Method")
		}
	}
}
