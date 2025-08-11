package usuario

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UsuarioHandler struct {
	service UsuarioService
}

func NewUsuarioHandler(s UsuarioService) *UsuarioHandler {
	return &UsuarioHandler{
		service: s,
	}
}

func getEmpresaIDFromContext(c *gin.Context) (uint, error) {
	valorEmpresaID, existe := c.Get("empresaID")
	if !existe {
		return 0, errors.New("ID da empresa não encontrado no contexto")
	}

	empresaIDStr, ok := valorEmpresaID.(string)
	if !ok {
		return 0, errors.New("ID da empresa no contexto está em formato inválido")
	}

	empresaID, err := strconv.ParseUint(empresaIDStr, 10, 64)
	if err != nil {
		return 0, errors.New("ID da empresa no token é inválido")
	}

	return uint(empresaID), nil
}

func (h *UsuarioHandler) GetByIdHandler(c *gin.Context) {
	empresaID, err := getEmpresaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID do usuário deve ser um número"})
		return
	}
	usuario, err := h.service.FindByID(uint(id), empresaID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}
	c.JSON(http.StatusOK, usuario)
}

func (h *UsuarioHandler) GetAllUsuariosHandler(c *gin.Context) {
	empresaID, err := getEmpresaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	usuarios, err := h.service.GetAll(empresaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar usuários"})
		return
	}
	c.JSON(http.StatusOK, usuarios)
}

func (h *UsuarioHandler) DeleteHandler(c *gin.Context) {
	empresaID, err := getEmpresaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	idUrl := c.Param("id")
	idToken, _ := c.Get("userID")
	if idUrl != idToken.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para deletar este usuário"})
		return
	}
	id, err := strconv.ParseUint(idUrl, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID do usuário deve ser um número"})
		return
	}

	err = h.service.Delete(uint(id), empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao deletar o usuário"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UsuarioHandler) UpdateUsuarioHandler(c *gin.Context) {
	empresaID, err := getEmpresaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	idToken, _ := c.Get("userID")
	idStr := c.Param("id")
	if idStr != idToken.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para editar este usuário"})
		return
	}

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

	err = h.service.Update(uint(id), empresaID, dadosParaAtualizar)
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
		Nome      string `json:"nome" binding:"required"`
		Email     string `json:"email" binding:"required,email"`
		Senha     string `json:"senha" binding:"required,min=6"`
		EmpresaID uint   `json:"empresa_id" binding:"required"`
		CargoID   uint   `json:"cargo_id" binding:"required"`
	}
	var request criarUsuarioRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	usuario := model.Usuario{
		Nome:      request.Nome,
		Email:     request.Email,
		Senha:     request.Senha,
		EmpresaID: request.EmpresaID,
		CargoID:   request.CargoID,
	}

	err := h.service.CriarUsuario(&usuario)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, usuario)
}

func (h *UsuarioHandler) GetMeuPerfil(c *gin.Context) {
	empresaID, err := getEmpresaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	valorID, _ := c.Get("userID")
	idString, _ := valorID.(string)
	id, _ := strconv.ParseUint(idString, 10, 64)

	usuario, err := h.service.FindByID(uint(id), empresaID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	c.JSON(http.StatusOK, usuario)
}
