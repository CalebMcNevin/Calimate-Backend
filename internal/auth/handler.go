package auth

import (
	"log"
	"net/http"

	"qc_api/internal/utils"

	"github.com/labstack/echo/v4"
)

// LoginHandler godoc
// @Summary User login
// @Description Authenticate user with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginDTO true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /login [post]
func (s *AuthService) LoginHandler(c echo.Context) error {
	var creds LoginDTO
	if err := c.Bind(&creds); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid JSON"})
	}
	if err := c.Validate(&creds); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
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

// RegisterHandler godoc
// @Summary User registration
// @Description Register a new user with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body RegisterDTO true "Registration credentials"
// @Success 200 {object} User
// @Failure 400 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /register [post]
func (s *AuthService) RegisterHandler(c echo.Context) error {
	var creds RegisterDTO
	if err := c.Bind(&creds); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid JSON"})
	}
	if err := c.Validate(&creds); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
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
