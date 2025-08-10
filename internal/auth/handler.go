package auth

import (
	"log"
	"net/http"

	"qc_api/internal/utils"

	"github.com/labstack/echo/v4"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *AuthService) LoginHandler(c echo.Context) error {
	var creds LoginRequest
	if err := c.Bind(&creds); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid JSON",
		})
	}

	user, err := s.AuthenticateUserPass(creds.Username, creds.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "Invalid credentials"})
	}
	token, err := s.GenerateJWT(user)
	if err != nil {
		log.Printf("Tried to create token: %v", err)
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Token generation failed"})
	}

	return c.JSON(http.StatusOK, LoginResponse{Token: token})
}

func (s *AuthService) RegisterHandler(c echo.Context) error {
	var creds RegisterRequest
	if err := c.Bind(&creds); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error: "invalid JSON",
		})
	}
	if len(creds.Username) < 4 {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "username must be at least 4 characters long"})
	}
	if len(creds.Password) < 8 {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "password must be at least 8 characters long"})
	}
	user, err := NewUser(creds.Username, creds.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	err = s.CreateUser(user)
	if err != nil {
		if err == ErrUserAlreadyExists {
			return c.JSON(http.StatusConflict, utils.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, user)
}
