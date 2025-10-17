package permissao

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/cache"
	"gorm.io/gorm"
)

type Repository interface {
	Create(permissao *model.Permissao) error
	FindAll() ([]model.Permissao, error)
	FindByCargoID(ctx context.Context, cargoID uint) ([]model.Permissao, error)
	AddPermissionToCargo(cargoID uint, permissaoID uint) error
}

type repository struct {
	Db    *gorm.DB
	cache cache.Service
}

func NewRepository(db *gorm.DB, cacheSvc cache.Service) Repository {
	return &repository{Db: db, cache: cacheSvc}
}

func (r *repository) Create(permissao *model.Permissao) error {
	return r.Db.Create(permissao).Error
}

func (r *repository) FindAll() ([]model.Permissao, error) {
	var permissoes []model.Permissao
	err := r.Db.Find(&permissoes).Error
	return permissoes, err
}

// FindByCargoID retorna permissões ligadas a um cargo especifico, com cache-aside.
func (r *repository) FindByCargoID(ctx context.Context, cargoID uint) ([]model.Permissao, error) {
	cacheKey := fmt.Sprintf("permissoes:cargo:%d", cargoID)
	if r.cache != nil {
		if cached, err := r.cache.Get(ctx, cacheKey); err == nil && cached != "" {
			var list []model.Permissao
			if errUM := json.Unmarshal([]byte(cached), &list); errUM == nil {
				log.Printf("[cache] HIT %s", cacheKey)
				return list, nil
			} else {
				log.Printf("[cache] falha ao desserializar %s: %v", cacheKey, errUM)
			}
		} else if err != nil {
			log.Printf("[cache] erro ao buscar chave %s: %v", cacheKey, err)
		}
	}
	var cargo model.Cargo
	if err := r.Db.Preload("Permissoes").First(&cargo, cargoID).Error; err != nil {
		return nil, err
	}
	log.Printf("[cache] MISS %s (carregado do DB)", cacheKey)
	if r.cache != nil {
		if b, mErr := json.Marshal(&cargo.Permissoes); mErr == nil {
			// TTL de 1h
			if sErr := r.cache.Set(ctx, cacheKey, string(b), time.Hour); sErr != nil {
				log.Printf("[cache] erro ao setar chave %s: %v", cacheKey, sErr)
			}
		} else {
			log.Printf("[cache] erro ao serializar %s: %v", cacheKey, mErr)
		}
	}
	return cargo.Permissoes, nil
}

// AddPermissionToCargo adiciona a permissão a um cargo e invalida o cache relacionado.
func (r *repository) AddPermissionToCargo(cargoID uint, permissaoID uint) error {
	var cargo model.Cargo
	var perm model.Permissao
	if err := r.Db.First(&cargo, cargoID).Error; err != nil {
		return err
	}
	if err := r.Db.First(&perm, permissaoID).Error; err != nil {
		return err
	}
	if err := r.Db.Model(&cargo).Association("Permissoes").Append(&perm); err != nil {
		return err
	}
	if r.cache != nil {
		key := fmt.Sprintf("permissoes:cargo:%d", cargoID)
		if err := r.cache.Delete(context.Background(), key); err != nil {
			log.Printf("[cache] erro ao invalidar %s após AddPermissionToCargo: %v", key, err)
		}
	}
	return nil
}
