package ponto

import (
	"context"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type RegistroPontoRepository interface {
	SavePonto(ctx context.Context, ponto *model.RegistroPonto) error
	FindPontosByUserIDAndDate(ctx context.Context, userID uint, dia time.Time) ([]model.RegistroPonto, error)
	FindPontosByUserIDAndDateRange(ctx context.Context, userID uint, inicio, fim time.Time) ([]model.RegistroPonto, error)
	WithTransaction(tx *gorm.DB) RegistroPontoRepository
	FindPontoByID(ctx context.Context, pontoID uint, empresaID uint) (*model.RegistroPonto, error)
	UpdatePonto(ctx context.Context, ponto *model.RegistroPonto) error
}

type pontoRepository struct {
	Db *gorm.DB
}

func NewPontoRepository(db *gorm.DB) RegistroPontoRepository {
	return &pontoRepository{Db: db}
}

func (r *pontoRepository) SavePonto(ctx context.Context, ponto *model.RegistroPonto) error {
	return r.Db.WithContext(ctx).Create(ponto).Error
}

func (r *pontoRepository) FindPontosByUserIDAndDate(ctx context.Context, userID uint, dia time.Time) ([]model.RegistroPonto, error) {
	ano, mes, diaDoMes := dia.Date()
	inicioDoDia := time.Date(ano, mes, diaDoMes, 0, 0, 0, 0, dia.Location())
	fimDoDia := time.Date(ano, mes, diaDoMes, 23, 59, 59, 0, dia.Location())

	var pontos []model.RegistroPonto
	err := r.Db.WithContext(ctx).Where("usuario_id = ?", userID).
		Where("timestamp BETWEEN ? AND ?", inicioDoDia, fimDoDia).
		Find(&pontos).Error
	return pontos, err
}

func (r *pontoRepository) FindPontosByUserIDAndDateRange(ctx context.Context, userID uint, inicio, fim time.Time) ([]model.RegistroPonto, error) {
	// Normalizar para garantir que inicio <= fim e remover nanos para consistência
	if fim.Before(inicio) {
		inicio, fim = fim, inicio
	}
	inicio = inicio.Truncate(time.Second)
	fim = fim.Truncate(time.Second)

	var pontos []model.RegistroPonto
	err := r.Db.WithContext(ctx).Where("usuario_id = ?", userID).
		Where("timestamp BETWEEN ? AND ?", inicio, fim).
		Order("timestamp ASC").
		Find(&pontos).Error
	return pontos, err
}

func (r *pontoRepository) WithTransaction(tx *gorm.DB) RegistroPontoRepository {
	return &pontoRepository{Db: tx}
}

func (r *pontoRepository) FindPontoByID(ctx context.Context, pontoID uint, empresaID uint) (*model.RegistroPonto, error) {
	var ponto model.RegistroPonto
	err := r.Db.WithContext(ctx).Where("id = ? AND empresa_id = ?", pontoID, empresaID).First(&ponto).Error
	if err != nil {
		return nil, err
	}
	return &ponto, nil
}

func (r *pontoRepository) UpdatePonto(ctx context.Context, ponto *model.RegistroPonto) error {
	return r.Db.WithContext(ctx).Save(ponto).Error
}
