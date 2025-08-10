package auth

import (
	"errors"
	"fmt"
	"log"
	"qc_api/internal/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	DB        *gorm.DB
	jwtSecret []byte
}

func NewAuthService(db *gorm.DB, jwtSecret []byte) *AuthService {
	return &AuthService{
		DB:        db,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) CreateUser(user *User) error {
	if err := s.DB.Create(user).Error; err != nil {
		if utils.IsUniqueConstraintError(err) {
			return ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

func (s *AuthService) getUserByUsername(username string) (*User, error) {
	var user User
	if err := s.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("User not found")
		}
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) getUserByID(userId string) (*User, error) {
	var user User
	if err := s.DB.First(&user, "id = ?", userId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("User not found")
		}
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) AuthenticateUserPass(username, password string) (*User, error) {
	user, err := s.getUserByUsername(username)

	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !checkPassword([]byte(user.HashedPassword), password) {
		return nil, ErrInvalidCredentials
	}
	if !checkPassword([]byte(user.HashedPassword), password) {
		return nil, ErrInvalidCredentials
	}
	return user, nil

}

func (s *AuthService) AuthenticateJWT(token string) (*User, error) {
	userId, err := s.validateJWT(token)
	if err != nil {
		return nil, err
	}

	return s.getUserByID(userId)
}

func hashPassword(password string) ([]byte, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v\n", err)
		return nil, err
	}
	return hashed, nil
}

func checkPassword(hashed []byte, password string) bool {
	return bcrypt.CompareHashAndPassword(hashed, []byte(password)) == nil
}

func (s *AuthService) GenerateJWT(user *User) (string, error) {
	claims := jwt.MapClaims{
		"username": user.Username,
		"user_id":  user.ID,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) validateJWT(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		// We only use HS256, so we check that the signing method is what we expect.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		return "", fmt.Errorf("invalid claims")
	}

	return claims["user_id"].(string), nil
}
