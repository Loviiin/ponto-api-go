package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/constants"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/ratelimit"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	authService  Service
	loginLimiter *ratelimit.RateLimiter
}

func NewAuthHandler(service Service) *AuthHandler {
	return &AuthHandler{
		authService: service,
		// SECURITY: Rate limiter: máximo 5 tentativas a cada 15 minutos
		loginLimiter: ratelimit.NewRateLimiter(5, 15*time.Minute),
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"superadmin@ponto.com"`
	Password string `json:"password" binding:"required" example:"superadmin"`
}

type DemoLoginResponse struct {
	Mensagem string `json:"mensagem"`
	Token    string `json:"token"`
	Email    string `json:"email"`
	UsuarioID uint   `json:"usuario_id"`
	EmpresaID uint   `json:"empresa_id"`
	CargoID   uint   `json:"cargo_id"`
	Permissoes []string `json:"permissoes"`
}

type SignUpRequest struct {
	Empresa struct {
		NomeFantasia string `json:"nome_fantasia" binding:"required" example:"Minha Empresa"`
		RazaoSocial  string `json:"razao_social" binding:"required" example:"Minha Empresa LTDA"`
		CNPJ         string `json:"cnpj" binding:"required" example:"12345678000195"`
	} `json:"empresa"`
	Localidade struct {
		Nome               string  `json:"nome" binding:"required" example:"Matriz Principal"`
		CEP                string  `json:"cep" binding:"required" example:"01001-000"`
		Logradouro         string  `json:"logradouro" example:"Rua Exemplo"`
		Bairro             string  `json:"bairro" example:"Centro"`
		Cidade             string  `json:"cidade" example:"São Paulo"`
		Estado             string  `json:"estado" example:"SP"`
		Latitude           float64 `json:"latitude" binding:"required" example:"-23.550520"`
		Longitude          float64 `json:"longitude" binding:"required" example:"-46.633308"`
		RaioGeofenceMetros float64 `json:"raio_geofence_metros" binding:"required" example:"100"`
	} `json:"localidade"`
	Usuario struct {
		Nome         string    `json:"nome" binding:"required" example:"João Administrador"`
		Email        string    `json:"email" binding:"required,email" example:"joao@empresa.com"`
		CPF          string    `json:"cpf" binding:"required" example:"12345678900"`
		Password     string    `json:"password" binding:"required,min=6" example:"senha123"`
		Salario      float64   `json:"salario" binding:"required" example:"5000.00"`
		DataAdmissao time.Time `json:"data_admissao" binding:"required" example:"2025-01-20T00:00:00Z"`
	} `json:"usuario"`
}

// @Summary      Realiza o login do usuário
// @Description  Autentica um usuário com email e senha e retorna um token JWT.
// @Tags         Autenticação
// @Accept       json
// @Produce      json
// @Param        loginRequest  body      LoginRequest  true  "Credenciais de Login"
// @Success      200           {object}  map[string]string
// @Failure      400           {object}  map[string]string
// @Failure      401           {object}  map[string]string
// @Failure      429           {object}  map[string]string
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"erro": err.Error()})
		return
	}

	// SECURITY: Rate limiting por email para prevenir brute force
	allowed, timeUntilRetry := h.loginLimiter.IsAllowed(request.Email)
	if !allowed {
		minutosRestantes := int(timeUntilRetry.Minutes()) + 1
		loc, _ := time.LoadLocation(constants.TimezoneBR)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"erro":        "Muitas tentativas de login falhadas. Tente novamente em " + time.Now().In(loc).Add(timeUntilRetry).Format("15:04:05"),
			"retry_after": minutosRestantes,
		})
		return
	}

	authenticate, err := h.authService.Authenticate(request.Email, request.Password)
	if err != nil {
		c.JSON(401, gin.H{"erro": err.Error()})
		return
	}

	// LOGIN BEM-SUCEDIDO: Resetar rate limiter
	h.loginLimiter.Reset(request.Email)
	c.JSON(200, gin.H{"token": authenticate})
}

// @Summary      Restaura e autentica o usuário demo
// @Description  Reseta o tenant demo para um estado limpo e autentica com credenciais hardcoded.
// @Tags         Autenticação
// @Produce      json
// @Success      200 {object} DemoLoginResponse
// @Failure      500 {object} map[string]string
// @Router       /auth/demo [post]
func (h *AuthHandler) Demo(c *gin.Context) {
	session, err := h.authService.PrepareDemoSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Falha ao preparar demo: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, DemoLoginResponse{
		Mensagem:   session.Mensagem,
		Token:      session.Token,
		Email:      session.Email,
		UsuarioID:  session.UsuarioID,
		EmpresaID:  session.EmpresaID,
		CargoID:    session.CargoID,
		Permissoes: session.Permissoes,
	})
}

// @Summary      Realiza o cadastro de uma nova empresa e seu administrador
// @Description  Cria uma nova empresa, a sua localidade principal (matriz), e o primeiro usuário administrador em uma única transação. Retorna o novo usuário e um token JWT.
// @Tags         Autenticação
// @Accept       json
// @Produce      json
// @Param        signUpRequest body      SignUpRequest true "Dados completos da Empresa, Localidade e Administrador" // CORRIGIDO: Swagger atualizado
// @Success      201           {object}  map[string]interface{}
// @Failure      400           {object}  map[string]string
// @Failure      500           {object}  map[string]string
// @Router       /auth/signup [post]
func (h *AuthHandler) SignUp(c *gin.Context) {
	var request SignUpRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Corpo da requisição inválido: " + err.Error()})
		return
	}

	empresa := &model.Empresa{
		NomeFantasia: request.Empresa.NomeFantasia,
		RazaoSocial:  request.Empresa.RazaoSocial,
		CNPJ:         request.Empresa.CNPJ,
	}

	localidade := &model.Localidade{
		Nome:               request.Localidade.Nome,
		CEP:                request.Localidade.CEP,
		Logradouro:         request.Localidade.Logradouro,
		Bairro:             request.Localidade.Bairro,
		Cidade:             request.Localidade.Cidade,
		Estado:             request.Localidade.Estado,
		Latitude:           request.Localidade.Latitude,
		Longitude:          request.Localidade.Longitude,
		RaioGeofenceMetros: request.Localidade.RaioGeofenceMetros,
	}

	usuario := &model.Usuario{
		Nome:  request.Usuario.Nome,
		Email: request.Usuario.Email,
		Senha: request.Usuario.Password,
		CPF:   request.Usuario.CPF,
	}

	dadosContrato := &model.Contrato{
		Salario:      request.Usuario.Salario,
		DataAdmissao: request.Usuario.DataAdmissao,
	}

	novoUsuario, token, err := h.authService.SignUp(empresa, localidade, usuario, dadosContrato)

	if err != nil {
		if strings.Contains(err.Error(), "e-mail já cadastrado") || strings.Contains(err.Error(), "CNPJ já cadastrado") {
			c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Falha ao realizar o cadastro: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"mensagem": "Empresa e usuário administrador criados com sucesso!",
		"usuario":  novoUsuario,
		"token":    token,
	})
}

// --- Password Reset Handlers ---

type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// @Summary      Solicita recuperação de senha
// @Description  Envia um email com link de recuperação de senha para o usuário.
// @Tags         Autenticação
// @Accept       json
// @Produce      json
// @Param        request body RequestPasswordResetRequest true "Email do usuário"
// @Success      200 {object} map[string]string
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req RequestPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	if err := h.authService.RequestPasswordReset(req.Email); err != nil {
		// Logar erro interno, mas retornar sucesso para não enumerar usuários
		// logger.Error("Erro ao solicitar reset de senha", err)
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Se o email estiver cadastrado, você receberá instruções para recuperar sua senha."})
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// @Summary      Redefine a senha
// @Description  Redefine a senha do usuário usando o token recebido por email.
// @Tags         Autenticação
// @Accept       json
// @Produce      json
// @Param        request body ResetPasswordRequest true "Token e nova senha"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Senha redefinida com sucesso."})
}

// @Summary      Valida token de recuperação
// @Description  Verifica se o token é válido e retorna o nome do usuário.
// @Tags         Autenticação
// @Accept       json
// @Produce      json
// @Param        token query string true "Token de recuperação"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Router       /auth/validate-reset-token [get]
func (h *AuthHandler) ValidateResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Token não fornecido"})
		return
	}

	user, err := h.authService.ValidateResetToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"nome":  user.Nome,
	})
}

// --- Google OAuth Handlers ---

// @Summary      Inicia login com Google
// @Description  Redireciona o usuário para a página de login do Google.
// @Tags         Autenticação
// @Success      302
// @Router       /auth/google/login [get]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := GetGoogleOAuthConfig().AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// @Summary      Callback do login com Google
// @Description  Recebe o código do Google e autentica o usuário.
// @Tags         Autenticação
// @Param        code query string true "Código de autorização do Google"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Router       /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Código não fornecido"})
		return
	}

	token, user, isNew, err := h.authService.AuthenticateWithGoogle(code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":  token,
		"user":   user,
		"is_new": isNew,
	})
}

// @Summary      Vincula conta Google (Início)
// @Description  Inicia o fluxo para vincular uma conta Google ao usuário logado.
// @Tags         Perfil
// @Security     BearerAuth
// @Success      302
// @Router       /auth/google/link [get]
func (h *AuthHandler) LinkGoogleAccount(c *gin.Context) {
	// Adicionar userID ao state para saber quem está vinculando no callback
	// Por simplicidade, vamos usar apenas um state fixo agora, mas idealmente seria um JWT assinado com o userID
	// Para MVP, o frontend vai chamar o callback passando o code para um endpoint autenticado
	url := GetGoogleOAuthConfig().AuthCodeURL("link-account", oauth2.AccessTypeOffline)
	c.JSON(http.StatusOK, gin.H{"url": url})
}

type LinkGoogleCallbackRequest struct {
	Code string `json:"code" binding:"required"`
}

// @Summary      Confirma vínculo de conta Google
// @Description  Recebe o código do Google e vincula ao usuário logado.
// @Tags         Perfil
// @Security     BearerAuth
// @Param        request body LinkGoogleCallbackRequest true "Código do Google"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Router       /auth/google/link/callback [post]
func (h *AuthHandler) LinkGoogleCallback(c *gin.Context) {
	var req LinkGoogleCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	userID := c.GetUint("userID") // Assumindo middleware de auth
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário não autenticado"})
		return
	}

	if err := h.authService.LinkGoogleAccount(userID, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Conta Google vinculada com sucesso!"})
}
