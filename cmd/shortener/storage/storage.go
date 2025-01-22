package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// URLRecord represents a single URL record
type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// URLStorage represents the URL storage
type URLStorage struct {
	urlmap map[string]URLRecord
	file   string
	mu     sync.Mutex
	nextID int // Следующий UUID
}

// NewURLStorage creates a new URL storage instance
func NewURLStorage(filePath string) *URLStorage {
	storage := &URLStorage{
		urlmap: make(map[string]URLRecord),
		file:   filePath,
		nextID: 1,
	}
	storage.loadFromFile() // Восстановление данных при старте
	return storage
}

// postURL adds a new short URL and original URL to the storage
func (u *URLStorage) PostURL(short, original string) string {
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

// getURL retrieves the original URL based on the short URL
func (u *URLStorage) GetURL(short string) (string, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	record, exists := u.urlmap[short]
	return record.OriginalURL, exists
}

// saveToFile saves the URL records to a file
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

	for _, record := range u.urlmap {
		data, err := json.Marshal(record)
		if err != nil {
			log.Printf("Failed to marshal record: %v\n", err)
			continue
		}
		// Записываем каждую строку в файл
		_, err = file.Write(append(data, '\n')) // Добавляем перенос строки
		if err != nil {
			log.Printf("Failed to write record to file: %v\n", err)
			return
		}
	}
}

// loadFromFile loads URL records from a file
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

	u.urlmap = make(map[string]URLRecord) // Сбрасываем текущие данные
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
}
