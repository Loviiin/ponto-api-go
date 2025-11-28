package model

import (
	"time"
)

type PasswordResetToken struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UsuarioID uint       `gorm:"index;not null" json:"usuario_id"`
	Token     string     `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
