package model

import (
	"time"

	"gorm.io/gorm"
)

type Chatbox struct {
	ID        uint            `json:"id" gorm:"not null"`
	AccountID uint            `json:"account_id" gorm:"not null"`
	Code      string          `json:"code" gorm:"not null;size:255;unique"`
	Name      string          `json:"name" gorm:"not null;size:255"`
	CreatedBy string          `json:"created_by" gorm:"size:255;default:SYSTEM"`
	UpdatedBy string          `json:"updated_by" gorm:"size:255;default:SYSTEM"`
	DeletedBy *string         `json:"deleted_by" gorm:"size:255"`
	CreatedAt *time.Time      `json:"created_at" gorm:"default:current_timestamp"`
	UpdatedAt *time.Time      `json:"updated_at" gorm:"default:current_timestamp"`
	DeletedAt *gorm.DeletedAt `json:"deleted_at"`
}
