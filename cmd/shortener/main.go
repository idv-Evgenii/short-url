package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/idv-Evgenii/short-url/cmd/shortener/config"
	"github.com/idv-Evgenii/short-url/cmd/shortener/handler"
	"github.com/idv-Evgenii/short-url/cmd/shortener/storage"
)

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
