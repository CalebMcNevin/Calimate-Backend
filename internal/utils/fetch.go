package utils

import (
	"fmt"
	"io"
	"net/http"
	"qc_api/internal/config"

	"github.com/labstack/gommon/log"
)

func GetMotiveUser(username string, cfg *config.Config) {

	url := "https://api.gomotive.com/v1/users/lookup?username=" + username

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("x-api-key", cfg.MotiveKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err.Error())
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading motive request body: %v", err)
	}

	fmt.Println(string(body))
}
