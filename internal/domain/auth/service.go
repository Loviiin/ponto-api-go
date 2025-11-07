package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/config"
	"github.com/Loviiin/ponto-api-go/internal/domain/cargo"
	"github.com/Loviiin/ponto-api-go/internal/domain/contrato" // NOVO IMPORT
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/domain/localidade" // NOVO IMPORT
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/cache"
	"github.com/Loviiin/ponto-api-go/pkg/geolocation"
	"github.com/Loviiin/ponto-api-go/pkg/jwt"
	"github.com/Loviiin/ponto-api-go/pkg/password"
	"github.com/Loviiin/ponto-api-go/pkg/validator"
	"gorm.io/gorm"
)

type AuthService interface {
	Authenticate(email string, password string) (string, error)
	SignUp(
		empresaReq *model.Empresa,
		localidadeParcial *model.Localidade,
		usuarioReq *model.Usuario,
		dadosContrato *model.Contrato,
	) (*model.Usuario, string, error)
}

type authService struct {
	usuarioRepo    usuario.UsuarioRepository
	empresaRepo    empresa.EmpresaRepository
	cargoRepo      cargo.CargoRepository
	contratoRepo   contrato.ContratoRepository
	localidadeRepo localidade.Repository
	geolocationSvc geolocation.Service
	jwtService     *jwt.JWTService
	cache          cache.Service // CACHE ADICIONADO
	db             *gorm.DB
}

func NewAuthService(
	usuarioRepo usuario.UsuarioRepository,
	empresaRepo empresa.EmpresaRepository,
	cargoRepo cargo.CargoRepository,
	contratoRepo contrato.ContratoRepository,
	localidadeRepo localidade.Repository,
	geolocationSvc geolocation.Service,
	jwtService *jwt.JWTService,
	cacheService cache.Service, // CACHE ADICIONADO
	db *gorm.DB,
) AuthService {
	return &authService{
		usuarioRepo:    usuarioRepo,
		empresaRepo:    empresaRepo,
		cargoRepo:      cargoRepo,
		contratoRepo:   contratoRepo,
		localidadeRepo: localidadeRepo,
		geolocationSvc: geolocationSvc,
		jwtService:     jwtService,
		cache:          cacheService, // CACHE ATRIBUÍDO
		db:             db,
	}
}

func (s *authService) Authenticate(email string, passwordStr string) (string, error) {
	ctx := context.Background()

	// SECURITY FIX: Normalizar email (lowercase + trim) para evitar problemas case-sensitive
	email = password.NormalizarEmail(email)

	// 1. CACHE: Verificar se usuário já está em cache (após login bem-sucedido)
	cacheKey := fmt.Sprintf("auth:user:%s", email)

	var usuari *model.Usuario

	// Tentar buscar do cache primeiro
	if cachedData, err := s.cache.Get(ctx, cacheKey); err == nil && cachedData != "" {
		var cachedUser model.Usuario
		if err := json.Unmarshal([]byte(cachedData), &cachedUser); err == nil {
			// Verificar senha mesmo com cache (segurança)
			if password.VerificaHashSenha(passwordStr, cachedUser.Senha) {
				// Cache HIT! Gerar token e retornar (economia de 1 query ao DB)
				token, err := s.jwtService.GenerateToken(cachedUser.ID, cachedUser.Contrato.EmpresaID)
				if err != nil {
					return "", err
				}
				return token, nil
			}
			// Senha incorreta - invalidar cache
			s.cache.Delete(ctx, cacheKey)
		}
	}

	// Cache MISS - buscar do banco de dados
	usuari, err := s.usuarioRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("credenciais inválidas")
		}
		return "", err
	}

	if usuari.Contrato.ID == 0 {
		return "", errors.New("utilizador não possui um contrato ativo")
	}

	if !password.VerificaHashSenha(passwordStr, usuari.Senha) {
		return "", errors.New("credenciais inválidas")
	}

	// 2. CACHE: Armazenar dados do usuário após login bem-sucedido
	if userData, err := json.Marshal(usuari); err == nil {
		// Cachear por 15 minutos (logins frequentes do mesmo usuário)
		s.cache.Set(ctx, cacheKey, string(userData), 15*time.Minute)
	}

	// 3. CACHE: Pré-carregar dados de perfil (antecipação)
	go s.precacheUserProfile(usuari.ID, usuari.Contrato.EmpresaID)

	// O token agora é gerado com o EmpresaID do Contrato
	token, err := s.jwtService.GenerateToken(usuari.ID, usuari.Contrato.EmpresaID)
	// --- FIM DA CORREÇÃO ---
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *authService) SignUp(
	empresaReq *model.Empresa,
	localidadeParcial *model.Localidade,
	usuarioReq *model.Usuario,
	dadosContrato *model.Contrato,
) (*model.Usuario, string, error) {
	// SECURITY FIX: Normalizar email antes de qualquer validação
	usuarioReq.Email = password.NormalizarEmail(usuarioReq.Email)

	// SECURITY FIX: Validar força da senha
	if err := password.ValidarForcaSenha(usuarioReq.Senha); err != nil {
		return nil, "", err
	}

	// DATA VALIDATION: Sanitizar e validar CPF
	usuarioReq.CPF = validator.SanitizarCPF(usuarioReq.CPF)
	if err := validator.ValidarCPF(usuarioReq.CPF); err != nil {
		return nil, "", err
	}

	// DATA VALIDATION: Sanitizar e validar CNPJ
	empresaReq.CNPJ = validator.SanitizarCNPJ(empresaReq.CNPJ)
	if err := validator.ValidarCNPJ(empresaReq.CNPJ); err != nil {
		return nil, "", err
	}

	_, err := s.usuarioRepo.FindByEmail(usuarioReq.Email)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", errors.New("e-mail já cadastrado")
	}

	// Adicionar verificação de CNPJ aqui se necessário

	var novoUsuario *model.Usuario
	var token string

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Criar a Empresa
		if err := s.empresaRepo.WithTransaction(tx).CreateEmpresa(empresaReq); err != nil {
			return fmt.Errorf("falha ao criar empresa: %w", err)
		}

		// 2. Validar coordenadas enviadas pelo frontend
		if localidadeParcial.Latitude < -90 || localidadeParcial.Latitude > 90 {
			return errors.New("latitude inválida (deve estar entre -90 e 90)")
		}
		if localidadeParcial.Longitude < -180 || localidadeParcial.Longitude > 180 {
			return errors.New("longitude inválida (deve estar entre -180 e 180)")
		}

		// Log para debug: verificar coordenadas recebidas
		fmt.Printf("📍 [SignUp] Coordenadas recebidas do frontend: lat=%.15f, lng=%.15f\n",
			localidadeParcial.Latitude, localidadeParcial.Longitude)

		// 3. Se coordenadas não foram enviadas, buscar do CEP como fallback
		if localidadeParcial.Latitude == 0 && localidadeParcial.Longitude == 0 {
			localidadeCompleta, err := s.geolocationSvc.GetLocationFromCEP(localidadeParcial.CEP)
			if err != nil {
				return fmt.Errorf("falha ao obter dados do CEP: %w", err)
			}
			// Preservar dados enviados pelo frontend e preencher apenas o que falta
			localidadeCompleta.Nome = localidadeParcial.Nome
			localidadeCompleta.RaioGeofenceMetros = localidadeParcial.RaioGeofenceMetros
			localidadeCompleta.EmpresaID = empresaReq.ID
			localidadeParcial = localidadeCompleta
		} else {
			// ✅ CORREÇÃO DO BUG: Usar as coordenadas enviadas pelo frontend
			// O usuário ajustou a posição no mapa, então respeitamos essa escolha
			localidadeParcial.EmpresaID = empresaReq.ID
		}

		if err := s.localidadeRepo.WithTransaction(tx).Save(localidadeParcial); err != nil {
			return fmt.Errorf("falha ao criar localidade: %w", err)
		}

		// 4. Configurar Cargos e Permissões e obter os cargos padrão
		permissoes := config.SeedPermissions(tx)
		donoCargo, _, _ := config.SetupDefaultRolesAndPermissions(tx, empresaReq.ID, permissoes)

		// 5. Preparar os dados do utilizador antes de o criar
		// Nota: Empresa e Cargo agora pertencem ao Contrato; o usuário não possui mais esses campos.

		senhaHash, err := password.CriptografaSenha(usuarioReq.Senha)
		if err != nil {
			return err
		}
		usuarioReq.Senha = senhaHash
		if err := s.usuarioRepo.WithTransaction(tx).Save(usuarioReq); err != nil {
			return fmt.Errorf("falha ao criar utilizador: %w", err)
		}

		// 5. Criar o Contrato
		novoContrato := &model.Contrato{
			UsuarioID:    usuarioReq.ID,
			EmpresaID:    empresaReq.ID,
			LocalidadeID: localidadeParcial.ID,
			CargoID:      donoCargo.ID,
			Salario:      dadosContrato.Salario,
			DataAdmissao: dadosContrato.DataAdmissao,
		}
		if err := s.contratoRepo.WithTransaction(tx).Save(novoContrato); err != nil {
			return fmt.Errorf("falha ao criar contrato: %w", err)
		}

		// 6. Gerar Token
		token, err = s.jwtService.GenerateToken(usuarioReq.ID, empresaReq.ID)
		if err != nil {
			return err
		}
		novoUsuario = usuarioReq
		novoUsuario.Contrato = *novoContrato

		return nil
	})

	return novoUsuario, token, err
}

// precacheUserProfile pré-carrega dados do perfil em background (melhora performance do GET /profile/me)
func (s *authService) precacheUserProfile(userID, empresaID uint) {
	ctx := context.Background()

	// 1. Cachear permissões do cargo (usado em middleware)
	permCacheKey := fmt.Sprintf("auth:permissions:user:%d", userID)

	// Buscar usuário com relacionamentos via DB direto (mais simples aqui)
	var user model.Usuario
	if err := s.db.Preload("Contrato.Cargo.Permissoes").First(&user, userID).Error; err == nil {
		if user.Contrato.ID > 0 && user.Contrato.Cargo.ID > 0 {
			if permData, err := json.Marshal(user.Contrato.Cargo.Permissoes); err == nil {
				s.cache.Set(ctx, permCacheKey, string(permData), 30*time.Minute)
			}
		}
	}

	// 2. Cachear dados da empresa (usado em várias telas)
	empCacheKey := fmt.Sprintf("auth:empresa:%d", empresaID)
	if empresa, err := s.empresaRepo.FindByID(ctx, empresaID); err == nil {
		if empData, err := json.Marshal(empresa); err == nil {
			s.cache.Set(ctx, empCacheKey, string(empData), 1*time.Hour)
		}
	}

	// 3. Cachear localidades da empresa (usado no GET /profile/me)
	locCacheKey := fmt.Sprintf("auth:localidades:empresa:%d", empresaID)
	var localidades []model.Localidade
	if err := s.db.Where("empresa_id = ?", empresaID).Find(&localidades).Error; err == nil {
		if locData, err := json.Marshal(localidades); err == nil {
			s.cache.Set(ctx, locCacheKey, string(locData), 30*time.Minute)
		}
	}
}

// InvalidateUserCache invalida todo cache relacionado a um usuário (usar após atualização de perfil)
func (s *authService) InvalidateUserCache(userID uint, email string, empresaID uint) {
	ctx := context.Background()

	// Invalidar todos os caches relacionados
	s.cache.Delete(ctx, fmt.Sprintf("auth:user:%s", email))
	s.cache.Delete(ctx, fmt.Sprintf("auth:permissions:user:%d", userID))
	s.cache.Delete(ctx, fmt.Sprintf("profile:user:%d", userID))
	s.cache.Delete(ctx, fmt.Sprintf("auth:empresa:%d", empresaID))
}
