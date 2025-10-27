package justificativa

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
	Create(justificativa *model.Justificativa) error
	FindByID(id uint, empresaID uint) (*model.Justificativa, error)
	FindByStatus(empresaID uint, status string) ([]model.Justificativa, error)
	FindByUsuarioID(usuarioID uint, empresaID uint) ([]model.Justificativa, error)
	FindByUsuarioIDAndPeriodo(usuarioID uint, empresaID uint, inicio, fim time.Time) ([]model.Justificativa, error)
	Update(justificativa *model.Justificativa) error
	InvalidarCacheEmpresa(empresaID uint)
	InvalidarCacheUsuario(usuarioID uint, empresaID uint)
	WithTransaction(tx *gorm.DB) Repository
}

type repository struct {
	Db    *gorm.DB
	cache cache.Service
}

func NewRepository(db *gorm.DB, cacheSvc cache.Service) Repository {
	return &repository{Db: db, cache: cacheSvc}
}

func (r *repository) Create(justificativa *model.Justificativa) error {
	if err := r.Db.Create(justificativa).Error; err != nil {
		return err
	}
	// Invalidar cache após criação
	r.InvalidarCacheEmpresa(justificativa.EmpresaID)
	r.InvalidarCacheUsuario(justificativa.UsuarioID, justificativa.EmpresaID)
	return nil
}

func (r *repository) WithTransaction(tx *gorm.DB) Repository {
	return &repository{Db: tx, cache: r.cache}
}

func (r *repository) FindByID(id uint, empresaID uint) (*model.Justificativa, error) {
	var justificativa model.Justificativa
	err := r.Db.Where("id = ? AND empresa_id = ?", id, empresaID).First(&justificativa).Error
	if err != nil {
		return nil, err
	}
	return &justificativa, nil
}

func (r *repository) FindByStatus(empresaID uint, status string) ([]model.Justificativa, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("justificativas:status:%s:empresa:%d", status, empresaID)

	// Tenta buscar do cache (apenas para status PENDENTE que muda menos)
	if r.cache != nil && status == "PENDENTE" {
		if cached, err := r.cache.Get(ctx, cacheKey); err == nil && cached != "" {
			var justificativas []model.Justificativa
			if errUM := json.Unmarshal([]byte(cached), &justificativas); errUM == nil {
				log.Printf("[cache] HIT justificativas pendentes empresa %d", empresaID)
				return justificativas, nil
			}
		}
	}

	log.Printf("[cache] MISS justificativas status %s empresa %d", status, empresaID)

	var justificativas []model.Justificativa
	err := r.Db.Where("empresa_id = ? AND status = ?", empresaID, status).
		Preload("Usuario").            // Carrega dados do solicitante
		Preload("Aprovador").          // Carrega dados do aprovador (se houver)
		Order("data_ocorrencia DESC"). // Mais recentes primeiro
		Find(&justificativas).Error

	// Armazena no cache (apenas pendentes, 10 minutos)
	if r.cache != nil && status == "PENDENTE" && err == nil {
		if b, mErr := json.Marshal(&justificativas); mErr == nil {
			_ = r.cache.Set(ctx, cacheKey, string(b), 10*time.Minute)
		}
	}

	return justificativas, err
}

func (r *repository) FindByUsuarioID(usuarioID uint, empresaID uint) ([]model.Justificativa, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("justificativas:usuario:%d:empresa:%d", usuarioID, empresaID)

	// Tenta buscar do cache
	if r.cache != nil {
		if cached, err := r.cache.Get(ctx, cacheKey); err == nil && cached != "" {
			var justificativas []model.Justificativa
			if errUM := json.Unmarshal([]byte(cached), &justificativas); errUM == nil {
				log.Printf("[cache] HIT justificativas usuário %d", usuarioID)
				return justificativas, nil
			}
		}
	}

	log.Printf("[cache] MISS justificativas usuário %d", usuarioID)

	var justificativas []model.Justificativa
	err := r.Db.Where("usuario_id = ? AND empresa_id = ?", usuarioID, empresaID).
		Preload("Usuario").
		Preload("Aprovador").
		Order("data_ocorrencia DESC").
		Find(&justificativas).Error

	// Armazena no cache (30 minutos)
	if r.cache != nil && err == nil {
		if b, mErr := json.Marshal(&justificativas); mErr == nil {
			_ = r.cache.Set(ctx, cacheKey, string(b), 30*time.Minute)
		}
	}

	return justificativas, err
}

func (r *repository) Update(justificativa *model.Justificativa) error {
	if err := r.Db.Save(justificativa).Error; err != nil {
		return err
	}
	// Invalidar cache após atualização (aprovação/reprovação)
	r.InvalidarCacheEmpresa(justificativa.EmpresaID)
	r.InvalidarCacheUsuario(justificativa.UsuarioID, justificativa.EmpresaID)
	return nil
}

func (r *repository) FindByUsuarioIDAndPeriodo(usuarioID uint, empresaID uint, inicio, fim time.Time) ([]model.Justificativa, error) {
	var justificativas []model.Justificativa
	err := r.Db.Where("usuario_id = ? AND empresa_id = ? AND data_ocorrencia BETWEEN ? AND ?",
		usuarioID, empresaID, inicio, fim).
		Find(&justificativas).Error
	return justificativas, err
}

// InvalidarCacheEmpresa invalida cache de listas da empresa (pendentes)
func (r *repository) InvalidarCacheEmpresa(empresaID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()

	// Invalida lista de pendentes
	keyPendentes := fmt.Sprintf("justificativas:status:PENDENTE:empresa:%d", empresaID)
	if err := r.cache.Delete(ctx, keyPendentes); err != nil {
		log.Printf("[cache] erro ao invalidar %s: %v", keyPendentes, err)
	}

	log.Printf("[cache] Invalidado cache de justificativas da empresa %d", empresaID)
}

// InvalidarCacheUsuario invalida cache de justificativas de um usuário específico
func (r *repository) InvalidarCacheUsuario(usuarioID uint, empresaID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()

	// Invalida lista do usuário
	keyUsuario := fmt.Sprintf("justificativas:usuario:%d:empresa:%d", usuarioID, empresaID)
	if err := r.cache.Delete(ctx, keyUsuario); err != nil {
		log.Printf("[cache] erro ao invalidar %s: %v", keyUsuario, err)
	}

	log.Printf("[cache] Invalidado cache de justificativas do usuário %d", usuarioID)
}
