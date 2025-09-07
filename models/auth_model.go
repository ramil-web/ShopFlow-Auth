package models

import (
	"time"
)

type User struct {
	ID         uint       `gorm:"primaryKey;autoIncrement"`
	Login      string     `gorm:"unique;not null"`
	Password   string     `gorm:"not null"`
	Email      string     `gorm:"unique;not null"`
	VerifiedAt *time.Time `gorm:"default:null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type UserResponse struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Login string `gorm:"unique;not null"`
	Email string `gorm:"unique;not null"`
}

type EmailMessage struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`
}
