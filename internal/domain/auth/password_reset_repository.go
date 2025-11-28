package auth

import (
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type PasswordResetRepository interface {
	Create(token *model.PasswordResetToken) error
	FindValidToken(tokenHash string) (*model.PasswordResetToken, error)
	MarkAsUsed(tokenID uint) error
	DeleteExpiredTokens() error
}

type passwordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

func (r *passwordResetRepository) Create(token *model.PasswordResetToken) error {
	return r.db.Create(token).Error
}

func (r *passwordResetRepository) FindValidToken(tokenHash string) (*model.PasswordResetToken, error) {
	var token model.PasswordResetToken
	err := r.db.Where("token = ? AND expires_at > ? AND used_at IS NULL", tokenHash, time.Now()).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *passwordResetRepository) MarkAsUsed(tokenID uint) error {
	now := time.Now()
	return r.db.Model(&model.PasswordResetToken{}).Where("id = ?", tokenID).Update("used_at", now).Error
}

func (r *passwordResetRepository) DeleteExpiredTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&model.PasswordResetToken{}).Error
}
