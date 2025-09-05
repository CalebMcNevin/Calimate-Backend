package auth

import (
	"qc_api/internal/db"
)

func Models() []any {
	return []any{
		&User{},
	}
}

type User struct {
	db.BaseModel
	Username       string `gorm:"uniqueIndex;not null" json:"username"`
	HashedPassword string `gorm:"not null"`
	Admin          bool   `gorm:"default:false"`
}

type LoginDTO struct {
	Username string `json:"username" validate:"required,min=4"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterDTO struct {
	Username string `json:"username" validate:"required,min=4"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func NewUser(username, password string) (*User, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &User{
		Username:       username,
		HashedPassword: string(hashedPassword),
	}
	return user, nil
}
