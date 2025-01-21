package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/idv-Evgenii/short-url/cmd/shortener/config"
	"github.com/idv-Evgenii/short-url/cmd/shortener/handler"
	"github.com/idv-Evgenii/short-url/cmd/shortener/storage"
)

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
	// Инициализация и настройки
	r := gin.Default()
	r.Use(handler.LoggerMiddleware())
	r.Use(handler.DecompressGzip())
	r.Use(handler.CompressGzip())

	// Получаем путь к файлу из флага или переменной окружения
	filePath := getFilePath()

	// Создаём хранилище с указанием файла
	storage := storage.NewURLStorage(filePath)

	// Конфигурация
	config := config.NewConfig()

	// Роутеры
	r.POST("/", handler.PostHandler(storage, config.BaseURL))
	r.GET("/:short", handler.PostHandler(storage, config.BaseURL))

	// Запуск сервера
	fmt.Printf("Listening on port %s\n", config.ServerAddress)
	r.Run(config.ServerAddress)
}
