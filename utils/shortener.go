package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateShortCode(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)[:n]
}
