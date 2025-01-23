package handler

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

func (m *MockStorage) PostURL(short, original string) string {
	m.data[short] = original
	return short
}

func (m *MockStorage) GetURL(short string) (string, bool) {
	val, exists := m.data[short]
	return val, exists
}

func TestPostHandler(t *testing.T) {
	storage := NewMockStorage()
	baseURL := "http://localhost:9090"

	r := gin.New()
	r.POST("/", PostHandler(storage, baseURL))

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
	_, err := buf.ReadFrom(res.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	responseBody := buf.String()

	if !strings.Contains(responseBody, baseURL+"/") {
		t.Errorf("expected response to contain '%s/', got '%s'", baseURL, responseBody)
	}
}

func TestAPIShortenHandler(t *testing.T) {
	storage := NewMockStorage()
	baseURL := "http://localhost:9090"

	r := gin.New()
	r.POST("/api/shorten", APIShortenHandler(storage, baseURL))

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
	original, found := storage.GetURL(short)
	if !found {
		t.Errorf("expected short URL '%s' to be saved", short)
	}
	if original != "https://example.com" {
		t.Errorf("expected original URL to be 'https://example.com', got '%s'", original)
	}
}
