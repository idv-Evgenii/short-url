package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
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

func (u *URLStorage) PostURL(short, original string) string {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Используем поле nextID, чтобы сгенерировать новый UUID
	uuid := fmt.Sprintf("%d", u.nextID)

	// Сохраняем запись в хранилище
	u.urlmap[short] = URLRecord{
		UUID:        uuid,
		ShortURL:    short,
		OriginalURL: original,
	}

	// Увеличиваем nextID для следующего вызова
	u.nextID++

	// Сохраняем данные в файл
	u.saveToFile()

	return uuid
}

func (u *URLStorage) GetURL(short string) (string, bool) {
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
