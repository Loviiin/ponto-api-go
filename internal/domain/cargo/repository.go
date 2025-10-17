package cargo

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

type CargoRepository interface {
	Create(cargo *model.Cargo) error
	FindByID(ctx context.Context, id uint, empresaID uint) (*model.Cargo, error)
	GetAllByEmpresaID(empresaID uint) ([]model.Cargo, error)
	Update(id uint, empresaID uint, dados map[string]interface{}) error
	Delete(id uint, empresaID uint) error
	AddPermissionToCargo(cargoID uint, permissaoID uint) error
	FindByName(nome string, empresaID uint) (*model.Cargo, error)
	WithTransaction(tx *gorm.DB) CargoRepository
}

type cargoRepository struct {
	Db    *gorm.DB
	cache cache.Service
}

// NewCargoRepository cria uma nova instância do repositório de cargos.
func NewCargoRepository(db *gorm.DB, cacheSvc cache.Service) CargoRepository {
	return &cargoRepository{Db: db, cache: cacheSvc}
}

func (r *cargoRepository) Create(cargo *model.Cargo) error {
	if err := r.Db.Create(cargo).Error; err != nil {
		return err
	}
	// Invalida lista de cargos da empresa
	if r.cache != nil {
		listKey := fmt.Sprintf("cargos:empresa:%d", cargo.EmpresaID)
		if err := r.cache.Delete(context.Background(), listKey); err != nil {
			log.Printf("[cache] erro ao invalidar %s após create: %v", listKey, err)
		}
	}
	return nil
}

func (r *cargoRepository) FindByID(ctx context.Context, id uint, empresaID uint) (*model.Cargo, error) {
	cacheKey := fmt.Sprintf("cargo:%d:empresa:%d", id, empresaID)

	if r.cache != nil {
		if cached, err := r.cache.Get(ctx, cacheKey); err == nil && cached != "" {
			var c model.Cargo
			if errUM := json.Unmarshal([]byte(cached), &c); errUM == nil {
				log.Printf("[cache] HIT %s", cacheKey)
				return &c, nil
			} else {
				log.Printf("[cache] falha ao desserializar cargo %d (empresa %d): %v", id, empresaID, errUM)
			}
		} else if err != nil {
			log.Printf("[cache] erro ao buscar chave %s: %v", cacheKey, err)
		}
	}

	var cargo model.Cargo
	err := r.Db.Preload("Permissoes").Where("id = ? AND empresa_id = ?", id, empresaID).First(&cargo).Error
	if err != nil {
		// Retorno antecipado para não popular cache com dados inválidos
		return nil, err
	}
	log.Printf("[cache] MISS %s (carregado do DB)", cacheKey)
	if r.cache != nil {
		if b, mErr := json.Marshal(&cargo); mErr == nil {
			if sErr := r.cache.Set(ctx, cacheKey, string(b), time.Hour); sErr != nil {
				log.Printf("[cache] erro ao setar chave %s: %v", cacheKey, sErr)
			}
		} else {
			log.Printf("[cache] erro ao serializar cargo %d (empresa %d): %v", id, empresaID, mErr)
		}
	}
	return &cargo, nil
}

func (r *cargoRepository) GetAllByEmpresaID(empresaID uint) ([]model.Cargo, error) {
	cacheKey := fmt.Sprintf("cargos:empresa:%d", empresaID)
	if r.cache != nil {
		if cached, err := r.cache.Get(context.Background(), cacheKey); err == nil && cached != "" {
			var list []model.Cargo
			if errUM := json.Unmarshal([]byte(cached), &list); errUM == nil {
				log.Printf("[cache] HIT %s", cacheKey)
				return list, nil
			} else {
				log.Printf("[cache] falha ao desserializar lista de cargos (empresa %d): %v", empresaID, errUM)
			}
		} else if err != nil {
			log.Printf("[cache] erro ao buscar chave %s: %v", cacheKey, err)
		}
	}

	var cargos []model.Cargo
	err := r.Db.Where("empresa_id = ?", empresaID).Order("id asc").Find(&cargos).Error
	if err != nil {
		return cargos, err
	}
	log.Printf("[cache] MISS %s (carregado do DB)", cacheKey)
	if r.cache != nil {
		if b, mErr := json.Marshal(&cargos); mErr == nil {
			if sErr := r.cache.Set(context.Background(), cacheKey, string(b), 5*time.Minute); sErr != nil {
				log.Printf("[cache] erro ao setar chave %s: %v", cacheKey, sErr)
			}
		} else {
			log.Printf("[cache] erro ao serializar lista de cargos (empresa %d): %v", empresaID, mErr)
		}
	}
	return cargos, nil
}

func (r *cargoRepository) Update(id uint, empresaID uint, dados map[string]interface{}) error {
	if err := r.Db.Model(&model.Cargo{}).Where("id = ? AND empresa_id = ?", id, empresaID).Updates(dados).Error; err != nil {
		return err
	}
	if r.cache != nil {
		// Invalida recurso e lista
		key := fmt.Sprintf("cargo:%d:empresa:%d", id, empresaID)
		if err := r.cache.Delete(context.Background(), key); err != nil {
			log.Printf("[cache] erro ao invalidar %s após update: %v", key, err)
		}
		listKey := fmt.Sprintf("cargos:empresa:%d", empresaID)
		if err := r.cache.Delete(context.Background(), listKey); err != nil {
			log.Printf("[cache] erro ao invalidar %s após update: %v", listKey, err)
		}
	}
	return nil
}

func (r *cargoRepository) Delete(id uint, empresaID uint) error {
	if err := r.Db.Unscoped().Delete(&model.Cargo{}, "id = ? AND empresa_id = ?", id, empresaID).Error; err != nil {
		return err
	}
	if r.cache != nil {
		key := fmt.Sprintf("cargo:%d:empresa:%d", id, empresaID)
		if err := r.cache.Delete(context.Background(), key); err != nil {
			log.Printf("[cache] erro ao invalidar %s após delete: %v", key, err)
		}
		listKey := fmt.Sprintf("cargos:empresa:%d", empresaID)
		if err := r.cache.Delete(context.Background(), listKey); err != nil {
			log.Printf("[cache] erro ao invalidar %s após delete: %v", listKey, err)
		}
	}
	return nil
}

func (r *cargoRepository) AddPermissionToCargo(cargoID uint, permissaoID uint) error {
	var cargo model.Cargo
	var permissao model.Permissao

	if err := r.Db.First(&cargo, cargoID).Error; err != nil {
		return err
	}
	if err := r.Db.First(&permissao, permissaoID).Error; err != nil {
		return err
	}
	return r.Db.Model(&cargo).Association("Permissoes").Append(&permissao)
}

func (r *cargoRepository) FindByName(nome string, empresaID uint) (*model.Cargo, error) {
	var cargo model.Cargo
	err := r.Db.Where("nome = ? AND empresa_id = ?", nome, empresaID).First(&cargo).Error
	if err != nil {
		return nil, err
	}
	return &cargo, nil
}

func (r *cargoRepository) WithTransaction(tx *gorm.DB) CargoRepository {
	return &cargoRepository{Db: tx, cache: r.cache}
}
