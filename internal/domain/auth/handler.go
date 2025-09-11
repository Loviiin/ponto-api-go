package auth

import (
	"net/http"
	"strings"

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
		Nome string `json:"nome" binding:"required" example:"Minha Empresa"`
	} `json:"empresa"`
	Usuario struct {
		Nome     string `json:"nome" binding:"required" example:"João Silva"`
		Email    string `json:"email" binding:"required,email" example:"joao@empresa.com"`
		Password string `json:"password" binding:"required" example:"senha123"`
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
// @Description  Cria uma nova empresa e o primeiro usuário administrador em uma única transação. Retorna o novo usuário e um token JWT.
// @Tags         Autenticação
// @Accept       json
// @Produce      json
// @Param        signUpRequest body      SignUpRequest true "Dados da Empresa e do Administrador"
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
        Nome: request.Empresa.Nome,
    }

    usuario := &model.Usuario{
        Nome:  request.Usuario.Nome,
        Email: request.Usuario.Email,
        Senha: request.Usuario.Password,
    }

    novoUsuario, token, err := h.authService.SignUp(empresa, usuario)
    if err != nil {
        if strings.Contains(err.Error(), "e-mail já cadastrado") {
             c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
             return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"erro": "Falha ao realizar o cadastro"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "mensagem": "Empresa e usuário administrador criados com sucesso!",
        "usuario":  novoUsuario,
        "token":    token,
    })
}