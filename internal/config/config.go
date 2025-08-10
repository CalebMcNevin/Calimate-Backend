package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
)

// Config holds the application configuration.
type Config struct {
	JWTSecret []byte
}

// NewConfig creates and returns a new configuration object.
func NewConfig() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		log.Println("!!! WARNING: JWT_SECRET environment variable not set.     !!!")
		log.Println("!!! Generating a temporary, insecure key for development. !!!")
		log.Println("!!! DO NOT use this in a production environment.          !!!")
		log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		jwtSecret = generateRandomKey(32)
	}

	return &Config{
		JWTSecret: []byte(jwtSecret),
	}
}

func generateRandomKey(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
