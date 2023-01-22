package models

import (
	"time"

	"gorm.io/gorm"
)

type UserToken struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserId    uint      `gorm:"not null" json:"user_id"`
	Type      string    `gorm:"size:50;not null;unique" json:"type"`
	Token     string    `gorm:"size:150;not null" json:"token"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `gorm:"foreignKey:UserId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
