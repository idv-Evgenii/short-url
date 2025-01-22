package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/idv-Evgenii/short-url/cmd/shortener/config"
	"github.com/idv-Evgenii/short-url/cmd/shortener/handler"
	"github.com/idv-Evgenii/short-url/cmd/shortener/storage"
)

func main() {
	r := gin.Default()
	r.Use(handler.LoggerMiddleware())
	r.Use(handler.DecompressGzip())
	r.Use(handler.CompressGzip())
	filePath := storage.GetFilePath()
	storage := storage.NewURLStorage(filePath)
	config := config.NewConfig()
	r.POST("/", handler.PostHandler(storage, config.BaseURL))
	r.GET("/:short", handler.PostHandler(storage, config.BaseURL))
	fmt.Printf("Listening on port %s\n", config.ServerAddress)
	if err := r.Run(config.ServerAddress); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}
