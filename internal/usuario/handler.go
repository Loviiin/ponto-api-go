package usuario

import (
	"errors"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)
import "github.com/gin-gonic/gin"

type UsuarioHandler struct {
	service UsuarioService
}

func NewUsuarioHandler(s UsuarioService) *UsuarioHandler {
	return &UsuarioHandler{
		service: s,
	}
}

func (h *UsuarioHandler) GetByIdHandler(c *gin.Context) {
	id := c.Param("id")

	convert, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID do usuário deve ser um número"})
		return
	}

	usuario, err := h.service.FindByID(uint(convert))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro dnv chefe": "a porra do id não existe"})
		return
	}
	c.JSON(http.StatusOK, usuario)
}

func (h *UsuarioHandler) GetAllUsuariosHandler(c *gin.Context) {
	usuarios, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro nessa porra": "erro na logica nessa porra chefe"})
		return
	}
	c.JSON(http.StatusOK, usuarios)

}

func (h *UsuarioHandler) UpdateUsuarioHandler(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID do usuário deve ser um número"})
		return
	}

	var dadosParaAtualizar map[string]interface{}
	if err := c.ShouldBindJSON(&dadosParaAtualizar); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corpo da requisição (JSON) inválido"})
		return
	}

	err = h.service.Update(uint(id), dadosParaAtualizar)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar o usuário"})
		return
	}

	c.Status(http.StatusNoContent)
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
