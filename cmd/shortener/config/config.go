package config

import (
	"flag"
	"os"
)

// Config структура для хранения конфигурации
type Config struct {
	ServerAddress string
	BaseURL       string
}

// getEnv получает значение переменной окружения с возможностью задать значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// NewConfig инициализирует конфигурацию
func NewConfig() *Config {
	// Флаги командной строки
	serverAddress := flag.String("a", "localhost:8080", "HTTP server address")
	baseURL := flag.String("b", "http://localhost:8080", "Base URL for short URLs")

	// Парсим флаги
	flag.Parse()

	// Приоритет: переменная окружения > флаг > значение по умолчанию
	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", *serverAddress),
		BaseURL:       getEnv("BASE_URL", *baseURL),
	}
}
