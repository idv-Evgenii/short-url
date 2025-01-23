package storage

import (
	"encoding/json"
	"flag"
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
	nextID int
}
type Storage interface {
	PostURL(short, original string) string
	GetURL(short string) (string, bool)
}

func NewURLStorage(filePath string) *URLStorage {
	storage := &URLStorage{
		urlmap: make(map[string]URLRecord),
		file:   filePath,
		nextID: 1,
	}
	storage.loadFromFile()
	return storage
}

func (u *URLStorage) PostURL(short, original string) string {
	u.mu.Lock()
	defer u.mu.Unlock()
	uuid := fmt.Sprintf("%d", u.nextID)
	u.urlmap[short] = URLRecord{
		UUID:        uuid,
		ShortURL:    short,
		OriginalURL: original,
	}
	u.nextID++
	u.saveToFile()
	return uuid
}

func (u *URLStorage) GetURL(short string) (string, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	record, exists := u.urlmap[short]
	return record.OriginalURL, exists
}

const filePermissions = 0644

func (u *URLStorage) saveToFile() {
	if u.file == "" {
		return
	}

	file, err := os.OpenFile(u.file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePermissions)
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
		_, err = file.Write(append(data, '\n'))
		if err != nil {
			log.Printf("Failed to write record to file: %v\n", err)
			return
		}
	}
}

func (u *URLStorage) loadFromFile() {
	if u.file == "" {
		return
	}
	file, err := os.Open(u.file)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to open file for loading: %v\n", err)
		}
		return
	}
	defer file.Close()

	u.urlmap = make(map[string]URLRecord)
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
func GetFilePath() string {
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
