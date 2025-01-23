package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/idv-Evgenii/short-url/cmd/shortener/storage"
)

const randomStringLength = 8

func PostHandler(u storage.Storage, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodPost:
			body, err := c.GetRawData()
			if err != nil || len(body) == 0 {
				c.String(http.StatusBadRequest, "Invalid Body")
				return
			}
			url := string(body)
			randomStr := getRandString(randomStringLength)
			_ = u.PostURL(randomStr, url)
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

func APIShortenHandler(u storage.Storage, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			URL string `json:"url"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		randomStr := getRandString(randomStringLength)
		_ = u.PostURL(randomStr, req.URL)
		c.JSON(http.StatusOK, gin.H{"result": baseURL + "/" + randomStr})
	}
}
