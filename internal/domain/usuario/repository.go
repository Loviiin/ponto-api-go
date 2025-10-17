package usuario

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

type UsuarioRepository interface {
	Save(usuario *model.Usuario) error
	FindByEmail(email string) (*model.Usuario, error)
	FindByID(ctx context.Context, id uint, empresaID uint) (*model.Usuario, error)
	GetAll(empresaID uint) ([]model.Usuario, error)
	Update(id uint, dados map[string]interface{}) error
	Delete(id uint) error
	FindAll() ([]model.Usuario, error)
	WithTransaction(tx *gorm.DB) UsuarioRepository
}

type usuarioRepository struct {
	Db    *gorm.DB
	cache cache.Service
}

func NewUsuarioRepository(db *gorm.DB, cacheSvc cache.Service) UsuarioRepository {
	return &usuarioRepository{Db: db, cache: cacheSvc}
}

func (r *usuarioRepository) Save(usuario *model.Usuario) error {
	if err := r.Db.Create(usuario).Error; err != nil {
		return err
	}
	if r.cache != nil {
		// Invalida listas: se conseguirmos obter o contrato recem-criado, invalidamos por empresa também
		var contrato model.Contrato
		if err := r.Db.Model(&model.Contrato{}).Select("empresa_id").Where("usuario_id = ?", usuario.ID).First(&contrato).Error; err == nil {
			listKey := fmt.Sprintf("usuarios:empresa:%d", contrato.EmpresaID)
			if err := r.cache.Delete(context.Background(), listKey); err != nil {
				log.Printf("[cache] erro ao invalidar %s após create: %v", listKey, err)
			}
		}
		if err := r.cache.Delete(context.Background(), "usuarios:all"); err != nil {
			log.Printf("[cache] erro ao invalidar 'usuarios:all' após create: %v", err)
		}
	}
	return nil
}

func (r *usuarioRepository) FindByEmail(email string) (*model.Usuario, error) {
	var usuario model.Usuario
	err := r.Db.Where("email = ?", email).
		Preload("Contrato.Cargo.Permissoes").
		First(&usuario).Error
	if err != nil {
		return nil, err
	}
	return &usuario, nil
}

func (r *usuarioRepository) FindByID(ctx context.Context, id uint, empresaID uint) (*model.Usuario, error) {
	// Cache key inclui empresa para evitar colisões cross-tenant
	cacheKey := fmt.Sprintf("usuario:%d:empresa:%d", id, empresaID)

	if r.cache != nil {
		if cached, err := r.cache.Get(ctx, cacheKey); err == nil && cached != "" {
			var u model.Usuario
			if errUM := json.Unmarshal([]byte(cached), &u); errUM == nil {
				log.Printf("[cache] HIT %s", cacheKey)
				return &u, nil
			} else {
				log.Printf("[cache] falha ao desserializar usuario %d (empresa %d): %v", id, empresaID, errUM)
			}
		} else if err != nil {
			log.Printf("[cache] erro ao buscar chave %s: %v", cacheKey, err)
		}
	}

	var usuario model.Usuario
	err := r.Db.Joins("JOIN contratos on contratos.usuario_id = usuarios.id").
		Where("usuarios.id = ? AND contratos.empresa_id = ?", id, empresaID).
		Preload("Contrato.Cargo.Permissoes").
		Preload("Contrato.Empresa").
		Preload("Contrato.Localidade").
		Preload("Contrato.Cargo.Permissoes").
		First(&usuario).Error
	if err != nil {
		// Retorno antecipado em caso de erro para evitar retornar struct não inicializado e não popular cache inválido
		return nil, err
	}

	// Apenas popula o cache quando a consulta for bem-sucedida
	log.Printf("[cache] MISS %s (carregado do DB)", cacheKey)
	if r.cache != nil {
		if b, mErr := json.Marshal(&usuario); mErr == nil {
			if sErr := r.cache.Set(ctx, cacheKey, string(b), 30*time.Minute); sErr != nil {
				log.Printf("[cache] erro ao setar chave %s: %v", cacheKey, sErr)
			}
		} else {
			log.Printf("[cache] erro ao serializar usuario %d (empresa %d): %v", id, empresaID, mErr)
		}
	}

	return &usuario, nil
}

func (r *usuarioRepository) GetAll(empresaID uint) ([]model.Usuario, error) {
	cacheKey := fmt.Sprintf("usuarios:empresa:%d", empresaID)
	if r.cache != nil {
		if cached, err := r.cache.Get(context.Background(), cacheKey); err == nil && cached != "" {
			var list []model.Usuario
			if errUM := json.Unmarshal([]byte(cached), &list); errUM == nil {
				log.Printf("[cache] HIT %s", cacheKey)
				return list, nil
			} else {
				log.Printf("[cache] falha ao desserializar lista de usuarios (empresa %d): %v", empresaID, errUM)
			}
		} else if err != nil {
			log.Printf("[cache] erro ao buscar chave %s: %v", cacheKey, err)
		}
	}

	var usuarios []model.Usuario
	err := r.Db.Joins("JOIN contratos ON contratos.usuario_id = usuarios.id").
		Where("contratos.empresa_id = ?", empresaID).
		Order("usuarios.id asc").
		Preload("Contrato.Cargo").
		Find(&usuarios).Error
	if err != nil {
		return usuarios, err
	}
	log.Printf("[cache] MISS %s (carregado do DB)", cacheKey)
	if r.cache != nil {
		if b, mErr := json.Marshal(&usuarios); mErr == nil {
			if sErr := r.cache.Set(context.Background(), cacheKey, string(b), 5*time.Minute); sErr != nil {
				log.Printf("[cache] erro ao setar chave %s: %v", cacheKey, sErr)
			}
		} else {
			log.Printf("[cache] erro ao serializar lista de usuarios (empresa %d): %v", empresaID, mErr)
		}
	}
	return usuarios, nil
}

func (r *usuarioRepository) Update(id uint, dados map[string]interface{}) error {
	if err := r.Db.Model(&model.Usuario{}).Where("id = ?", id).Updates(dados).Error; err != nil {
		return err
	}
	if r.cache != nil {
		// Tenta descobrir empresaID para invalidar cache do recurso e listas por empresa
		var contrato model.Contrato
		if err := r.Db.Model(&model.Contrato{}).Select("empresa_id").Where("usuario_id = ?", id).First(&contrato).Error; err == nil {
			key := fmt.Sprintf("usuario:%d:empresa:%d", id, contrato.EmpresaID)
			if err := r.cache.Delete(context.Background(), key); err != nil {
				log.Printf("[cache] erro ao invalidar %s após update: %v", key, err)
			}
			listKey := fmt.Sprintf("usuarios:empresa:%d", contrato.EmpresaID)
			if err := r.cache.Delete(context.Background(), listKey); err != nil {
				log.Printf("[cache] erro ao invalidar %s após update: %v", listKey, err)
			}
		}
		// Também invalida lista global
		if err := r.cache.Delete(context.Background(), "usuarios:all"); err != nil {
			log.Printf("[cache] erro ao invalidar 'usuarios:all' após update: %v", err)
		}
	}
	return nil
}

func (r *usuarioRepository) Delete(id uint) error {
	var empresaID uint
	// Buscar empresaID antes de apagar contratos (para invalidar caches por empresa)
	var contrato model.Contrato
	if err := r.Db.Model(&model.Contrato{}).Select("empresa_id").Where("usuario_id = ?", id).First(&contrato).Error; err == nil {
		empresaID = contrato.EmpresaID
	}
	if err := r.Db.Where("usuario_id = ?", id).Delete(&model.Contrato{}).Error; err != nil {
		return err
	}
	if err := r.Db.Unscoped().Delete(&model.Usuario{}, id).Error; err != nil {
		return err
	}
	if r.cache != nil {
		if empresaID != 0 {
			key := fmt.Sprintf("usuario:%d:empresa:%d", id, empresaID)
			if err := r.cache.Delete(context.Background(), key); err != nil {
				log.Printf("[cache] erro ao invalidar %s após delete: %v", key, err)
			}
			listKey := fmt.Sprintf("usuarios:empresa:%d", empresaID)
			if err := r.cache.Delete(context.Background(), listKey); err != nil {
				log.Printf("[cache] erro ao invalidar %s após delete: %v", listKey, err)
			}
		}
		if err := r.cache.Delete(context.Background(), "usuarios:all"); err != nil {
			log.Printf("[cache] erro ao invalidar 'usuarios:all' após delete: %v", err)
		}
	}
	return nil
}

func (r *usuarioRepository) FindAll() ([]model.Usuario, error) {
	var usuarios []model.Usuario
	err := r.Db.Preload("Contrato").Find(&usuarios).Error
	return usuarios, err
}

func (r *usuarioRepository) WithTransaction(tx *gorm.DB) UsuarioRepository {
	return &usuarioRepository{Db: tx, cache: r.cache}
}
