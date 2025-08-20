package ponto

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
	"time"
)

type RegistroPontoRepository interface {
	SavePonto(ponto *model.RegistroPonto) error
	FindPontosByUserIDAndDate(userID uint, dia time.Time) ([]model.RegistroPonto, error)
}

type pontoRepository struct {
	Db *gorm.DB
}

func NewPontoRepository(db *gorm.DB) RegistroPontoRepository {
	return &pontoRepository{Db: db}
}

func (r *pontoRepository) SavePonto(ponto *model.RegistroPonto) error {
	return r.Db.Create(ponto).Error
}

func (r *pontoRepository) FindPontosByUserIDAndDate(userID uint, dia time.Time) ([]model.RegistroPonto, error) {
	ano, mes, diaDoMes := dia.Date()
	inicioDoDia := time.Date(ano, mes, diaDoMes, 0, 0, 0, 0, dia.Location())
	fimDoDia := time.Date(ano, mes, diaDoMes, 23, 59, 59, 0, dia.Location())

	var pontos []model.RegistroPonto
	err := r.Db.Where("usuario_id = ?", userID).
		Where("timestamp BETWEEN ? AND ?", inicioDoDia, fimDoDia).
		Find(&pontos).Error
	return pontos, err
}
