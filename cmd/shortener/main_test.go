package main

import (
	"bytes"
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

func TestPostHandler(t *testing.T) {
	storage := NewMockStorage()
	handler := postHandler(storage)
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

	if !strings.Contains(responseBody, "http://localhost:8080/") {
		t.Errorf("expected response to contain 'http://localhost:8080/', got '%s'", responseBody)
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
func TestGetHandler(t *testing.T) {
	storage := NewMockStorage()
	storage.postURL("abc123", "https://example.com")
	handler := postHandler(storage)
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
func TestGetHandler_NotFound(t *testing.T) {
	storage := NewMockStorage()
	handler := postHandler(storage)

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
