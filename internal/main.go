package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/idv-Evgenii/short-url/cmd/shortener/config"
	"github.com/sirupsen/logrus"
)

// URLRecord represents a single URL record
type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type url interface {
	postURL(short, original string) string
	getURL(short string) (string, bool)
}

type URLStorage struct {
	urlmap map[string]URLRecord
	file   string
	mu     sync.Mutex
	nextID int // Следующий UUID
}

func NewURLStorage(filePath string) *URLStorage {
	storage := &URLStorage{
		urlmap: make(map[string]URLRecord),
		file:   filePath,
		nextID: 1,
	}
	storage.loadFromFile() // Восстановление данных при старте
	return storage
}

func (u *URLStorage) postURL(short, original string) string {
	u.mu.Lock()
	defer u.mu.Unlock()
	uuid := fmt.Sprintf("%d", u.nextID)
	u.urlmap[short] = URLRecord{
		UUID:        uuid,
		ShortURL:    short,
		OriginalURL: original,
	}
	u.nextID++     // Увеличиваем UUID для следующей записи
	u.saveToFile() // Сохранение после добавления
	return uuid
}

func (u *URLStorage) getURL(short string) (string, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	record, exists := u.urlmap[short]
	return record.OriginalURL, exists
}

func (u *URLStorage) saveToFile() {
	if u.file == "" {
		return // Если файл не задан, не сохраняем
	}
	file, err := os.OpenFile(u.file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("Failed to open file for saving: %v\n", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range u.urlmap {
		if err := encoder.Encode(record); err != nil {
			log.Printf("Failed to write to file: %v\n", err)
			return
		}
	}

	// Сохраняем следующий UUID
	state := map[string]int{"nextID": u.nextID}
	if err := encoder.Encode(state); err != nil {
		log.Printf("Failed to save state: %v\n", err)
	}
}

func (u *URLStorage) loadFromFile() {
	if u.file == "" {
		return // Если файл не задан, пропускаем загрузку
	}
	file, err := os.Open(u.file)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to open file for loading: %v\n", err)
		}
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		var record URLRecord
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Failed to decode JSON entry: %v\n", err)
			return
		}
		u.urlmap[record.ShortURL] = record
	}

	// Восстанавливаем следующий UUID
	var state map[string]int
	if err := decoder.Decode(&state); err == nil {
		u.nextID = state["nextID"]
	} else {
		log.Printf("Failed to restore nextID: %v\n", err)
	}
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

		// Сжимаем данные
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
			_ = u.postURL(randomStr, url) // Теперь uuid используется в методе postURL, но мы не выводим его напрямую
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

func getFilePath() string {
	defaultPath := "/tmp/short-url-db.json"
	filePath := flag.String("f", "", "Path to the file for URL storage")
	flag.Parse()

	envPath := os.Getenv("FILE_STORAGE_PATH")
	if envPath != "" {
		return envPath
	}

	if *filePath != "" {
		return *filePath
	}

	return defaultPath
}

func main() {
	r := gin.Default()
	r.Use(LoggerMiddleware())
	r.Use(decompressGzip())
	r.Use(compressGzip())

	// Получаем путь к файлу из флага или переменной окружения
	filePath := getFilePath()
	if filePath == "" {
		log.Println("File storage is disabled")
	}

	// Создаём хранилище с указанием файла
	storage := NewURLStorage(filePath)

	config := config.NewConfig()
	r.POST("/", postHandler(storage, config.BaseURL))
	r.GET("/:short", postHandler(storage, config.BaseURL))

	fmt.Printf("Listening on port %s\n", config.ServerAddress)
	r.Run(config.ServerAddress)
}
