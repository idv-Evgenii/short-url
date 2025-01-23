package storage

import (
	"os"
	"testing"
)

func TestPostURL(t *testing.T) {
	filePath := "/tmp/test_storage.json"
	defer os.Remove(filePath)

	storage := NewURLStorage(filePath)
	short := "short123"
	original := "https://example.com"

	storage.PostURL(short, original)

	result, exists := storage.GetURL(short)
	if !exists {
		t.Errorf("URL '%s' was not saved", short)
	}

	if result != original {
		t.Errorf("expected original URL to be '%s', got '%s'", original, result)
	}
}

func TestSaveAndLoadFromFile(t *testing.T) {
	filePath := "/tmp/test_storage.json"
	defer os.Remove(filePath)

	storage := NewURLStorage(filePath)
	short := "short123"
	original := "https://example.com"

	storage.PostURL(short, original)
	storage = NewURLStorage(filePath)

	result, exists := storage.GetURL(short)
	if !exists {
		t.Errorf("URL '%s' was not loaded from file", short)
	}

	if result != original {
		t.Errorf("expected original URL to be '%s', got '%s'", original, result)
	}
}
