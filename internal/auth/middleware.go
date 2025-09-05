package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *AuthService) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			if err := c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Missing token",
			}); err != nil {
				return err
			}
			return errors.New("missing auth token")
		}

		userId, err := s.validateJWT(token)
		if err != nil {
			if err := c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid token",
			}); err != nil {
				return err
			}
			return errors.New("invalid auth token")
		}
		parsedUserID, err := uuid.Parse(userId)
		if err != nil {
			return errors.New("failed to parse UserID")
		}
		c.Set("user_id", parsedUserID)
		return next(c)
	}
}
