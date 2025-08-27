package auth

import (
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
