package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type MockStorage struct {
	data map[string]string
}

func NewMockStorage() *MockStorage {
	return &MockStorage{data: make(map[string]string)}
}

func (m *MockStorage) postURL(short, original string) {
	m.data[short] = original
}

func (m *MockStorage) getURL(short string) (string, bool) {
	val, exists := m.data[short]
	return val, exists
}

// TestPostHandler проверяет основной POST-эндпоинт
func TestPostHandler(t *testing.T) {
	storage := NewMockStorage()
	baseURL := "http://localhost:9090"
	handler := postHandler(storage, baseURL)

	r := gin.New()
	r.POST("/", handler)

	body := bytes.NewBufferString("https://example.com")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	responseBody := buf.String()

	if !strings.Contains(responseBody, baseURL+"/") {
		t.Errorf("expected response to contain '%s/', got '%s'", baseURL, responseBody)
	}

	found := false
	for short, original := range storage.data {
		if original == "https://example.com" {
			if !strings.Contains(responseBody, short) {
				t.Errorf("expected short URL '%s' to be in response '%s'", short, responseBody)
			}
			found = true
		}
	}

	if !found {
		t.Error("expected URL 'https://example.com' to be saved in storage")
	}
}

// TestGetHandler проверяет основной GET-эндпоинт
func TestGetHandler(t *testing.T) {
	storage := NewMockStorage()
	storage.postURL("abc123", "https://example.com")
	baseURL := "http://localhost:9090"
	handler := postHandler(storage, baseURL)

	r := gin.New()
	r.GET("/:short", handler)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, res.StatusCode)
	}

	location := res.Header.Get("Location")
	if location != "https://example.com" {
		t.Errorf("expected Location header to be 'https://example.com', got '%s'", location)
	}
}

// TestGetHandler_NotFound проверяет, что некорректный GET-запрос возвращает 404
func TestGetHandler_NotFound(t *testing.T) {
	storage := NewMockStorage()
	baseURL := "http://localhost:9090"
	handler := postHandler(storage, baseURL)

	r := gin.New()
	r.GET("/:short", handler)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	responseBody := buf.String()

	if !strings.Contains(responseBody, "Not found Url") {
		t.Errorf("expected response to contain 'Not found Url', got '%s'", responseBody)
	}
}

// TestApiShortenHandler проверяет новый эндпоинт /api/shorten
func TestApiShortenHandler(t *testing.T) {
	storage := NewMockStorage()
	baseURL := "http://localhost:9090"
	handler := apiShortenHandler(storage, baseURL)

	r := gin.New()
	r.POST("/api/shorten", handler)

	// Тестовый JSON-запрос
	body := map[string]string{"url": "https://example.com"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	var resBody map[string]string
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	resultURL, exists := resBody["result"]
	if !exists || !strings.HasPrefix(resultURL, baseURL+"/") {
		t.Errorf("expected 'result' field to start with '%s/', got '%s'", baseURL, resultURL)
	}

	// Проверяем, что ссылка сохранена
	short := strings.TrimPrefix(resultURL, baseURL+"/")
	original, found := storage.getURL(short)
	if !found {
		t.Errorf("expected short URL '%s' to be saved", short)
	}
	if original != "https://example.com" {
		t.Errorf("expected original URL to be 'https://example.com', got '%s'", original)
	}
}
