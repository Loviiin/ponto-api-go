package handler

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/internal/service"
	"net/http"
)
import "github.com/gin-gonic/gin"

type UsuarioHandler struct {
	service service.UsuarioService
}

func NewUsuarioHandler(s service.UsuarioService) *UsuarioHandler {
	return &UsuarioHandler{
		service: s,
	}
}

func (h *UsuarioHandler) CriarUsuarioHandler(c *gin.Context) {

	type criarUsuarioRequest struct {
		Nome  string `json:"nome" binding:"required"`
		Email string `json:"email" binding:"required,email"`
		Senha string `json:"senha" binding:"required,min=6"`
		Cargo string `json:"cargo" binding:"required"`
	}
	var request criarUsuarioRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	usuario := model.Usuario{
		Nome:  request.Nome,
		Email: request.Email,
		Senha: request.Senha,
		Cargo: request.Cargo,
	}

	err := h.service.CriarUsuario(&usuario)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, usuario)
}
