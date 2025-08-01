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

func (h *AuthHandler) Login(c *gin.Context) {

	type LoginRequest struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

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
