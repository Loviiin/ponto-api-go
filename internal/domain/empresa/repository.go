package empresa

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

type EmpresaRepository interface {
	FindByID(ctx context.Context, id uint) (*model.Empresa, error)
	CreateEmpresa(empresa *model.Empresa) error
	GetAllEmpresas() ([]model.Empresa, error)
	GetEmpresaByID(idempresa uint) (*model.Empresa, error)
	UpdateEmpresa(idempresa uint, dados map[string]interface{}) error
	DeleteEmpresa(idempresa uint) error
	WithTransaction(tx *gorm.DB) EmpresaRepository
}

type empresaRepository struct {
	Db    *gorm.DB
	cache cache.Service
}

func NewEmpresaRepository(db *gorm.DB, cacheSvc cache.Service) EmpresaRepository {
	return &empresaRepository{Db: db, cache: cacheSvc}
}

func (r *empresaRepository) FindByID(ctx context.Context, id uint) (*model.Empresa, error) {
	// Cache key padronizada por recurso
	cacheKey := fmt.Sprintf("empresa:%d", id)

	// 1) Tenta buscar no cache
	if r.cache != nil {
		if cached, err := r.cache.Get(ctx, cacheKey); err == nil && cached != "" {
			var emp model.Empresa
			if errUM := json.Unmarshal([]byte(cached), &emp); errUM == nil {
				log.Printf("[cache] HIT %s", cacheKey)
				return &emp, nil
			} else {
				// Se falhar o unmarshal, loga e segue para o DB
				log.Printf("[cache] falha ao desserializar empresa %d: %v", id, errUM)
			}
		} else if err != nil {
			log.Printf("[cache] erro ao buscar chave %s: %v", cacheKey, err)
		}
	}

	// 2) Cache miss: busca no DB
	var empresa model.Empresa
	err := r.Db.Where("id = ?", id).First(&empresa).Error
	if err != nil {
		// Retorno antecipado para não popular cache com dados inválidos
		return nil, err
	}

	log.Printf("[cache] MISS %s (carregado do DB)", cacheKey)
	// 3) Popular o cache (1h)
	if r.cache != nil {
		if b, mErr := json.Marshal(&empresa); mErr == nil {
			if sErr := r.cache.Set(ctx, cacheKey, string(b), time.Hour); sErr != nil {
				log.Printf("[cache] erro ao setar chave %s: %v", cacheKey, sErr)
			}
		} else {
			log.Printf("[cache] erro ao serializar empresa %d: %v", id, mErr)
		}
	}

	return &empresa, nil
}

func (r *empresaRepository) CreateEmpresa(empresa *model.Empresa) error {
	if err := r.Db.Create(empresa).Error; err != nil {
		return err
	}
	// Invalida lista de empresas
	if r.cache != nil {
		if err := r.cache.Delete(context.Background(), "empresas:all"); err != nil {
			log.Printf("[cache] erro ao invalidar 'empresas:all' após create: %v", err)
		}
	}
	return nil
}

func (r *empresaRepository) GetAllEmpresas() ([]model.Empresa, error) {
	// Cache-aside para lista de empresas
	const cacheKey = "empresas:all"
	if r.cache != nil {
		if cached, err := r.cache.Get(context.Background(), cacheKey); err == nil && cached != "" {
			var empresasCached []model.Empresa
			if errUM := json.Unmarshal([]byte(cached), &empresasCached); errUM == nil {
				log.Printf("[cache] HIT %s", cacheKey)
				return empresasCached, nil
			} else {
				log.Printf("[cache] falha ao desserializar lista de empresas: %v", errUM)
			}
		} else if err != nil {
			log.Printf("[cache] erro ao buscar chave %s: %v", cacheKey, err)
		}
	}

	var empresas []model.Empresa
	err := r.Db.Order("id asc").Find(&empresas).Error
	if err != nil {
		return empresas, err
	}

	log.Printf("[cache] MISS %s (carregado do DB)", cacheKey)
	if r.cache != nil {
		if b, mErr := json.Marshal(&empresas); mErr == nil {
			if sErr := r.cache.Set(context.Background(), cacheKey, string(b), 5*time.Minute); sErr != nil {
				log.Printf("[cache] erro ao setar chave %s: %v", cacheKey, sErr)
			}
		} else {
			log.Printf("[cache] erro ao serializar lista de empresas: %v", mErr)
		}
	}

	return empresas, nil
}

func (r *empresaRepository) GetEmpresaByID(idempresa uint) (*model.Empresa, error) {
	var empresa model.Empresa
	err := r.Db.Where("id = ?", idempresa).First(&empresa).Error
	if err != nil {
		return nil, err
	}
	return &empresa, nil
}

func (r *empresaRepository) UpdateEmpresa(idempresa uint, dados map[string]interface{}) error {
	if err := r.Db.Model(&model.Empresa{}).Where("id = ?", idempresa).Updates(dados).Error; err != nil {
		return err
	}
	// Invalida cache do recurso e lista
	if r.cache != nil {
		if err := r.cache.Delete(context.Background(), fmt.Sprintf("empresa:%d", idempresa)); err != nil {
			log.Printf("[cache] erro ao invalidar empresa:%d após update: %v", idempresa, err)
		}
		if err := r.cache.Delete(context.Background(), "empresas:all"); err != nil {
			log.Printf("[cache] erro ao invalidar 'empresas:all' após update: %v", err)
		}
	}
	return nil
}

// DeleteEmpresa remove uma empresa do banco de dados.
func (r *empresaRepository) DeleteEmpresa(idempresa uint) error {
	// A função Unscoped() garante uma exclusão permanente (hard delete).
	// Sem ela, o GORM faria um soft delete se o modelo tivesse um campo gorm.DeletedAt.
	if err := r.Db.Unscoped().Delete(&model.Empresa{}, idempresa).Error; err != nil {
		return err
	}
	if r.cache != nil {
		if err := r.cache.Delete(context.Background(), fmt.Sprintf("empresa:%d", idempresa)); err != nil {
			log.Printf("[cache] erro ao invalidar empresa:%d após delete: %v", idempresa, err)
		}
		if err := r.cache.Delete(context.Background(), "empresas:all"); err != nil {
			log.Printf("[cache] erro ao invalidar 'empresas:all' após delete: %v", err)
		}
	}
	return nil
}

func (r *empresaRepository) WithTransaction(tx *gorm.DB) EmpresaRepository {
	return &empresaRepository{Db: tx, cache: r.cache}
}
