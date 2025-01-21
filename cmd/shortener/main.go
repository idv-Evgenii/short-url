package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/idv-Evgenii/short-url/cmd/shortener/config"
	"github.com/sirupsen/logrus"
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

func decompressGzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.String(http.StatusBadRequest, "Failed to decompress request: %v", err)
				c.Abort()
				return
			}
			defer gz.Close()

			body, err := io.ReadAll(gz)
			if err != nil {
				c.String(http.StatusBadRequest, "Failed to read decompressed body: %v", err)
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}
		c.Next()
	}
}
func compressGzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		writer := &responseCapture{
			ResponseWriter: c.Writer,
			buf:            bytes.NewBuffer(nil),
		}
		c.Writer = writer

		c.Next()

		if !supportsGzip(c) || (c.Writer.Header().Get("Content-Type") != "application/json" && c.Writer.Header().Get("Content-Type") != "text/html") {
			writer.ResponseWriter.Write(writer.buf.Bytes())
			return
		}
		var compressedBuf bytes.Buffer
		gz := gzip.NewWriter(&compressedBuf)
		_, err := gz.Write(writer.buf.Bytes())
		gz.Close()
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to compress response")
			return
		}
		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Header().Set("Content-Length", fmt.Sprint(compressedBuf.Len()))
		c.Writer.WriteHeader(c.Writer.Status())
		c.Writer.Write(compressedBuf.Bytes())
	}
}

type responseCapture struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (r *responseCapture) Write(data []byte) (int, error) {
	return r.buf.Write(data)
}

func supportsGzip(c *gin.Context) bool {
	return c.GetHeader("Accept-Encoding") != "" && c.GetHeader("Accept-Encoding") != "gzip"
}

func postHandler(u url, baseURL string) gin.HandlerFunc {
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
			c.String(http.StatusCreated, "%s/%s", baseURL, randomStr)
		case http.MethodGet:
			shortURL := c.Param("short")
			original, exists := u.getURL(shortURL)
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

func LoggerMiddleware() gin.HandlerFunc {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		contentLength := c.Writer.Size()

		logger.WithFields(logrus.Fields{
			"method":        c.Request.Method,
			"uri":           c.Request.RequestURI,
			"status":        statusCode,
			"contentLength": contentLength,
			"duration":      duration.String(),
		}).Info("Handled request")
	}
}

type RequestBody struct {
	URL string `json:"url"`
}

type ResponseBody struct {
	Result string `json:"result"`
}

func apiShortenHandler(u url, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqBody RequestBody
		if err := c.ShouldBindJSON(&reqBody); err != nil || reqBody.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		randomStr := getRandString(8)
		u.postURL(randomStr, reqBody.URL)
		resBody := ResponseBody{
			Result: fmt.Sprintf("%s/%s", baseURL, randomStr),
		}
		c.JSON(http.StatusOK, resBody)
	}
}

func main() {
	r := gin.Default()
	r.Use(LoggerMiddleware())
	r.Use(decompressGzip())
	r.Use(compressGzip())

	config := config.NewConfig()
	storage := NewURLStorage()
	r.POST("/", postHandler(storage, config.BaseURL))
	r.GET("/:short", postHandler(storage, config.BaseURL))
	r.POST("/api/shorten", apiShortenHandler(storage, config.BaseURL))

	fmt.Printf("Listening on port %s\n", config.ServerAddress)
	r.Run(config.ServerAddress)
}
