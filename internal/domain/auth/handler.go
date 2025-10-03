package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{
		authService: service,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"superadmin@ponto.com"`
	Password string `json:"password" binding:"required" example:"superadmin"`
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
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {

	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"erro": err.Error()})
		return
	}
	authenticate, err := h.authService.Authenticate(request.Email, request.Password)
	if err != nil {
		c.JSON(401, gin.H{"erro": err.Error()})
		return
	}
	c.JSON(200, gin.H{"token": authenticate})
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