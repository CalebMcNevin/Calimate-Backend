package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strconv"
)

// Config holds the application configuration.
type Config struct {
	JWTSecret   []byte
	MotiveKey   string
	AuthTimeout int
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

	authTokenTimeout, err := strconv.Atoi(os.Getenv("AUTH_TOKEN_TIMEOUT"))
	if err != nil {
		log.Printf("Error loading AUTH_TOKEN_TIMEOUT: %v\n", err.Error())
		authTokenTimeout = 60 * 60 * 1000
		log.Printf("No auth token timeout set (AUTH_TOKEN_TIMEOUT). Using default of %d\n", authTokenTimeout)
	}

	motiveKey := os.Getenv("MOTIVE_KEY")
	if motiveKey == "" {
		log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		log.Println("!!! WARNING: MOTIVE_KEY environment variable not set.     !!!")
		log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}

	return &Config{
		JWTSecret:   []byte(jwtSecret),
		MotiveKey:   motiveKey,
		AuthTimeout: authTokenTimeout,
	}
}

func generateRandomKey(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic("failed to generate random key")
	}
	return hex.EncodeToString(bytes)
}
