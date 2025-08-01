package handler

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/internal/service"
)
import "github.com/gin-gonic/gin"

type UsuarioHandler struct {
	service *service.UsuarioService
}

func NewUsuarioHandler(s *service.UsuarioService) *UsuarioHandler {
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
	erro := c.ShouldBindJSON(&request)
	if erro != nil {
		c.JSON(400, gin.H{"erro": erro.Error()})
		return
	}
	var Usuariodevdd = model.Usuario{}
	Usuariodevdd.Nome = request.Nome
	Usuariodevdd.Email = request.Email
	Usuariodevdd.Senha = request.Senha
	Usuariodevdd.Cargo = request.Cargo

	err := h.service.CriarUsuario(&Usuariodevdd)
	if err != nil {
		c.JSON(409, gin.H{"erro": err.Error()})
		return
	}
	c.JSON(201, Usuariodevdd)
}
