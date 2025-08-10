package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func (s *AuthService) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Missing token",
			})
			return errors.New("Missing auth token")
		}

		userId, err := s.validateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid token",
			})
			return errors.New("Invalid auth token")
		}
		c.Set("user_id", userId)
		return next(c)
	}
}
