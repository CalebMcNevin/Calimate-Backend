package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"qc_api/internal/auth"
	"qc_api/internal/config"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&auth.User{})
	return db
}

func TestRegister(t *testing.T) {
	// Setup
	cfg := config.NewConfig()
	db := setupTestDB()
	authService := auth.NewAuthService(db, cfg.JWTSecret)
	e := echo.New()

	// Request
	reqBody := `{"username": "testuser", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler
	err := authService.RegisterHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestLoginIncorrectPassword(t *testing.T) {
	// Setup
	cfg := config.NewConfig()
	db := setupTestDB()
	authService := auth.NewAuthService(db, cfg.JWTSecret)
	e := echo.New()

	// Create a user
	user, _ := auth.NewUser("testuser", "password123")
	authService.CreateUser(user)

	// Request
	reqBody := `{"username": "testuser", "password": "wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler
	err := authService.LoginHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware(t *testing.T) {
	// Setup
	cfg := config.NewConfig()
	db := setupTestDB()
	authService := auth.NewAuthService(db, cfg.JWTSecret)
	e := echo.New()

	// Create a user
	user, _ := auth.NewUser("testuser", "password123")
	authService.CreateUser(user)

	// Login to get a token
	token, _ := authService.GenerateJWT(user)

	// Request
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Middleware
	h := authService.AuthMiddleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Assertions
	assert.NoError(t, h(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}
