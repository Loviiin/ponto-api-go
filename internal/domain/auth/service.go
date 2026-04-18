package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"log/slog"

	"github.com/Loviiin/ponto-api-go/internal/config"
	"github.com/Loviiin/ponto-api-go/internal/domain/cargo"
	"github.com/Loviiin/ponto-api-go/internal/domain/contrato"
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/domain/localidade"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/cache"
	"github.com/Loviiin/ponto-api-go/pkg/email"
	"github.com/Loviiin/ponto-api-go/pkg/geolocation"
	"github.com/Loviiin/ponto-api-go/pkg/jwt"
	"github.com/Loviiin/ponto-api-go/pkg/password"
	"github.com/Loviiin/ponto-api-go/pkg/validator"
	oauth2_v2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

// Service define a interface para o serviço de autenticação
type Service interface {
	Authenticate(email, password string) (string, error)
	SignUp(empresaReq *model.Empresa, localidadeParcial *model.Localidade, usuarioReq *model.Usuario, dadosContrato *model.Contrato) (*model.Usuario, string, error)
	RequestPasswordReset(email string) error
	ResetPassword(token, newPassword string) error
	ResetDemoEnvironment() error
	LinkGoogleAccount(userID uint, code string) error
	AuthenticateWithGoogle(code string) (string, *model.Usuario, bool, error)
	InvalidateUserCache(userID uint, email string, empresaID uint)
	ValidateResetToken(token string) (*model.Usuario, error)
}

type authService struct {
	usuarioRepo       usuario.UsuarioRepository
	empresaRepo       empresa.EmpresaRepository
	cargoRepo         cargo.CargoRepository
	contratoRepo      contrato.ContratoRepository
	localidadeRepo    localidade.Repository
	passwordResetRepo PasswordResetRepository
	geolocationSvc    geolocation.Service
	jwtService        *jwt.JWTService
	cache             cache.Service
	db                *gorm.DB
	emailService      *email.EmailService
}

// NewAuthService cria uma nova instância do serviço de autenticação
func NewAuthService(
	usuarioRepo usuario.UsuarioRepository,
	empresaRepo empresa.EmpresaRepository,
	cargoRepo cargo.CargoRepository,
	contratoRepo contrato.ContratoRepository,
	localidadeRepo localidade.Repository,
	passwordResetRepo PasswordResetRepository,
	geolocationSvc geolocation.Service,
	jwtService *jwt.JWTService,
	cache cache.Service,
	db *gorm.DB,
	emailService *email.EmailService,
) Service {
	return &authService{
		usuarioRepo:       usuarioRepo,
		empresaRepo:       empresaRepo,
		cargoRepo:         cargoRepo,
		contratoRepo:      contratoRepo,
		localidadeRepo:    localidadeRepo,
		passwordResetRepo: passwordResetRepo,
		geolocationSvc:    geolocationSvc,
		jwtService:        jwtService,
		cache:             cache,
		db:                db,
		emailService:      emailService,
	}
}

func (s *authService) Authenticate(emailStr, passwordStr string) (string, error) {
	ctx := context.Background()

	// SECURITY FIX: Normalizar email (lowercase + trim) para evitar problemas case-sensitive
	emailStr = password.NormalizarEmail(emailStr)

	// 1. CACHE: Verificar se usuário já está em cache (após login bem-sucedido)
	cacheKey := fmt.Sprintf("auth:user:%s", emailStr)

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
	usuari, err := s.usuarioRepo.FindByEmail(emailStr)
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
	// Skip in test environment where DB might be nil
	if s.db == nil {
		return
	}

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

func (s *authService) ResetDemoEnvironment() error {
	if err := config.ResetAndSeedDemoWorkspace(s.db); err != nil {
		return err
	}

	if s.cache == nil {
		return nil
	}

	if flusher, ok := s.cache.(cache.Flusher); ok {
		if err := flusher.FlushAll(context.Background()); err != nil {
			return err
		}
	}

	return nil
}

// --- Password Reset & Google OAuth ---

func (s *authService) RequestPasswordReset(emailStr string) error {
	ctx := context.Background()
	emailStr = password.NormalizarEmail(emailStr)

	user, err := s.usuarioRepo.FindByEmail(emailStr)
	if err != nil {
		// Por segurança, não revelamos se o email existe ou não
		return nil
	}

	// Gerar token aleatório
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return err
	}
	token := hex.EncodeToString(randomBytes)

	// Hash do token para salvar no banco
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Salvar no banco
	resetToken := &model.PasswordResetToken{
		UsuarioID: user.ID,
		Token:     tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if err := s.passwordResetRepo.Create(resetToken); err != nil {
		return err
	}

	// Salvar no Cache (Upstash) para validação rápida
	cacheKey := fmt.Sprintf("password_reset:%s", tokenHash)
	tokenData := map[string]interface{}{
		"user_id":    user.ID,
		"email":      user.Email,
		"expires_at": resetToken.ExpiresAt.Unix(),
	}
	if data, err := json.Marshal(tokenData); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), 1*time.Hour)
	}

	// Enviar Email
	resetLink := s.emailService.BuildPasswordResetLink(token)
	// body := email.GetTemplate(email.TemplatePasswordReset)
	// Simplificação: substituir placeholder no template (idealmente usar html/template com struct)
	// Mas como o template usa {{.Link}}, podemos usar o EmailService.SendEmail que já faz isso

	data := struct {
		Link string
		Name string
		Year int
	}{
		Link: resetLink,
		Name: user.Nome,
		Year: time.Now().Year(),
	}

	// Enviar em goroutine para não bloquear
	go func() {
		if err := s.emailService.SendEmail([]string{user.Email}, "Recuperação de Senha - Nexora Ponto", email.TemplatePasswordReset, data); err != nil {
			slog.Error("Falha ao enviar email de reset", "error", err)
		}
	}()

	return nil
}

func (s *authService) ResetPassword(tokenRaw string, newPassword string) error {
	ctx := context.Background()

	// Validar força da senha
	if len(newPassword) < 6 {
		return errors.New("a senha deve ter no mínimo 6 caracteres")
	}

	// Hash do token recebido
	hash := sha256.Sum256([]byte(tokenRaw))
	tokenHash := hex.EncodeToString(hash[:])

	// 1. Tentar validar pelo Cache (Rápido)
	cacheKey := fmt.Sprintf("password_reset:%s", tokenHash)
	var userID uint

	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(cachedData), &data); err == nil {
			// Verificar expiração
			expiresAt := int64(data["expires_at"].(float64))
			if time.Now().Unix() > expiresAt {
				return errors.New("token expirado")
			}
			userID = uint(data["user_id"].(float64))
		}
	}

	// 2. Se não achou no cache ou erro, buscar no Banco (Fallback)
	if userID == 0 {
		tokenDB, err := s.passwordResetRepo.FindValidToken(tokenHash)
		if err != nil {
			return errors.New("token inválido ou expirado")
		}
		userID = tokenDB.UsuarioID

		// Marcar como usado no banco
		if err := s.passwordResetRepo.MarkAsUsed(tokenDB.ID); err != nil {
			return err
		}
	} else {
		// Se veio do cache, precisamos marcar como usado no banco também (para auditoria)
		// Mas como não temos o ID do token no cache, buscamos pelo hash
		if tokenDB, err := s.passwordResetRepo.FindValidToken(tokenHash); err == nil {
			s.passwordResetRepo.MarkAsUsed(tokenDB.ID)
		}
	}

	// 3. Atualizar senha do usuário
	hashedPassword, err := password.CriptografaSenha(newPassword)
	if err != nil {
		return err
	}

	// Atualizar senha
	if err := s.usuarioRepo.Update(userID, map[string]interface{}{"senha": hashedPassword}); err != nil {
		return err
	}

	// 4. Limpar Cache
	s.cache.Delete(ctx, cacheKey) // Remove token de reset

	// Invalidar cache de autenticação do usuário
	// Buscar email do usuário diretamente (sem precisar de empresa_id)
	var user model.Usuario
	if err := s.db.Select("email").Where("id = ?", userID).First(&user).Error; err == nil {
		// Invalidar cache de auth
		s.cache.Delete(ctx, fmt.Sprintf("auth:user:%s", user.Email))
		s.cache.Delete(ctx, fmt.Sprintf("auth:permissions:user:%d", userID))
		s.cache.Delete(ctx, fmt.Sprintf("profile:user:%d", userID))
	}

	return nil
}

func (s *authService) LinkGoogleAccount(userID uint, code string) error {
	// Trocar code por token
	conf := GetGoogleOAuthConfig()
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("falha ao trocar code por token: %v", err)
	}

	// Obter dados do usuário Google
	client := conf.Client(context.Background(), token)
	oauth2Service, err := oauth2_v2.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("falha ao criar serviço oauth2: %v", err)
	}

	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return fmt.Errorf("falha ao obter dados do usuário: %v", err)
	}

	// Buscar usuário atual
	// Precisamos do empresaID para o FindByID
	// Vamos assumir que o userID é válido e buscar contrato para saber empresa
	// Ou usar um método que busca só por ID? O repo atual é multi-tenant strict.
	// Vou usar FindAll ou algo assim? Não.
	// Vou adicionar um método FindByIDSimple no repo? Não posso alterar repo agora facilmente.
	// Vou tentar buscar contrato primeiro?

	var contrato model.Contrato
	if err := s.db.Where("usuario_id = ?", userID).First(&contrato).Error; err != nil {
		return err
	}

	user, err := s.usuarioRepo.FindByID(context.Background(), userID, contrato.EmpresaID)
	if err != nil {
		return err
	}

	// Buscar empresa para verificar estratégia
	empresa, err := s.empresaRepo.FindByID(context.Background(), user.Contrato.EmpresaID)
	if err != nil {
		return err
	}

	// Verificar estratégia
	if empresa.GoogleLinkStrategy == "strict" {
		if userInfo.Email != user.Email {
			return errors.New("email do Google não corresponde ao email da sua conta (Modo Estrito)")
		}
	} else {
		// Flexible
		if userInfo.Email != user.Email {
			// TODO: Implementar envio de email de confirmação
		}
	}

	// Salvar Google ID
	// user.GoogleID = &userInfo.Id // user é *model.Usuario
	// Mas Update pede map ou struct? Repo Update pede map.

	if err := s.usuarioRepo.Update(user.ID, map[string]interface{}{"google_id": userInfo.Id}); err != nil {
		return err
	}

	// Invalidar cache
	s.cache.Delete(context.Background(), fmt.Sprintf("auth:user:%s", user.Email))

	return nil
}

func (s *authService) AuthenticateWithGoogle(code string) (string, *model.Usuario, bool, error) {
	// Trocar code por token
	conf := GetGoogleOAuthConfig()
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		return "", nil, false, fmt.Errorf("falha ao trocar code por token: %v", err)
	}

	// Obter dados do usuário Google
	client := conf.Client(context.Background(), token)
	oauth2Service, err := oauth2_v2.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return "", nil, false, fmt.Errorf("falha ao criar serviço oauth2: %v", err)
	}

	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return "", nil, false, fmt.Errorf("falha ao obter dados do usuário: %v", err)
	}

	// Verificar se usuário existe pelo Google ID
	// Precisamos de um método FindByGoogleID no repo?
	// Ou buscar por email e verificar GoogleID?

	user, err := s.usuarioRepo.FindByEmail(userInfo.Email)
	if err != nil {
		return "", nil, false, errors.New("usuário não encontrado, entre em contato com o RH")
	}

	// Se usuário existe, verificar se tem GoogleID vinculado
	if user.GoogleID == nil {
		return "", nil, false, errors.New("conta Google não vinculada, faça login com senha e vincule no perfil")
	}

	if *user.GoogleID != userInfo.Id {
		return "", nil, false, errors.New("conta Google incorreta para este usuário")
	}

	// Gerar JWT
	jwtToken, err := s.jwtService.GenerateToken(user.ID, user.Contrato.EmpresaID)
	if err != nil {
		return "", nil, false, err
	}

	return jwtToken, user, false, nil
}

func (s *authService) ValidateResetToken(tokenRaw string) (*model.Usuario, error) {
	ctx := context.Background()

	// Hash do token recebido
	hash := sha256.Sum256([]byte(tokenRaw))
	tokenHash := hex.EncodeToString(hash[:])

	// 1. Tentar validar pelo Cache (Rápido)
	cacheKey := fmt.Sprintf("password_reset:%s", tokenHash)
	var userID uint

	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(cachedData), &data); err == nil {
			// Verificar expiração
			expiresAt := int64(data["expires_at"].(float64))
			if time.Now().Unix() > expiresAt {
				return nil, errors.New("token expirado")
			}
			userID = uint(data["user_id"].(float64))
		}
	}

	// 2. Se não achou no cache ou erro, buscar no Banco (Fallback)
	if userID == 0 {
		tokenDB, err := s.passwordResetRepo.FindValidToken(tokenHash)
		if err != nil {
			return nil, errors.New("token inválido ou expirado")
		}
		userID = tokenDB.UsuarioID
	}

	// 3. Buscar usuário para retornar nome
	// Precisamos do empresaID para o FindByID, mas aqui não temos fácil.
	// Vamos tentar buscar o contrato primeiro para pegar a empresaID
	var contrato model.Contrato
	if err := s.db.Where("usuario_id = ?", userID).First(&contrato).Error; err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	user, err := s.usuarioRepo.FindByID(ctx, userID, contrato.EmpresaID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
