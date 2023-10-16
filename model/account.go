package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Account struct {
	ID          uint            `json:"id" gorm:"not null"`
	Name        string          `json:"name" gorm:"not null;size:255"`
	Email       string          `json:"email" gorm:"size:255;unique"`
	PhoneNumber string          `json:"phone_number" gorm:"size:255;unique"`
	Password    string          `json:"password" gorm:"size:255"`
	Address     string          `json:"address" gorm:"size:255"`
	Type        string          `json:"type" gorm:"size:255"`
	ValidUntil  datatypes.Date  `json:"valid_until"`
	CreatedBy   string          `json:"created_by" gorm:"size:255;default:SYSTEM"`
	UpdatedBy   string          `json:"updated_by" gorm:"size:255;default:SYSTEM"`
	DeletedBy   *string         `json:"deleted_by" gorm:"size:255"`
	CreatedAt   *time.Time      `json:"created_at" gorm:"default:current_timestamp"`
	UpdatedAt   *time.Time      `json:"updated_at" gorm:"default:current_timestamp"`
	DeletedAt   *gorm.DeletedAt `json:"deleted_at"`
}
