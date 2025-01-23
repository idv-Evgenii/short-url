package handler

import (
	"crypto/rand"
	"encoding/hex"
)

func getRandString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err) // Выберите, как обработать ошибку.
	}
	return hex.EncodeToString(b)[:n]
}
