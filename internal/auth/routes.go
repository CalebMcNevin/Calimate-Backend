package auth

import "github.com/labstack/echo/v4"

func RegisterRoutes(e *echo.Echo, authService *AuthService) {
	e.POST("/login", authService.LoginHandler)
	e.POST("/register", authService.RegisterHandler)
}
