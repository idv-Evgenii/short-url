package config

import "flag"

// Config структура для хранения конфигурации
type Config struct {
	ServerAddress string
	BaseURL       string
}

// NewConfig инициализирует конфигурацию с аргументов командной строки
func NewConfig() *Config {
	serverAddress := flag.String("a", "localhost:8080", "HTTP server address")
	baseURL := flag.String("b", "http://localhost:8080", "Base URL for short URLs")

	flag.Parse()

	return &Config{
		ServerAddress: *serverAddress,
		BaseURL:       *baseURL,
	}
}
