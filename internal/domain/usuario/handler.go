package usuario

import (
	"errors"
	"github.com/Loviiin/ponto-api-go/internal/domain/cargo"
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type UsuarioHandler struct {
	service        UsuarioService
	empresaService empresa.EmpresaService
	cargoService   cargo.CargoService
	converter      funcoes.FuncoesInterface
}

func NewUsuarioHandler(s UsuarioService, e empresa.EmpresaService, ca cargo.CargoService, f funcoes.FuncoesInterface) *UsuarioHandler {
	return &UsuarioHandler{
		service:        s,
		empresaService: e,
		cargoService:   ca,
		converter:      f,
	}
}

func (h *UsuarioHandler) GetByIdHandler(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID do usuário deve ser um número"})
		return
	}
	usuario, err := h.service.FindByID(id, empresaID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}
	c.JSON(http.StatusOK, usuario)
}

func (h *UsuarioHandler) GetAllUsuariosHandler(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
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
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	idToken, err := h.converter.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	idUrl, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID do usuário deve ser um número"})
		return
	}

	if idUrl != idToken {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para deletar este usuário"})
		return
	}
	err = h.service.Delete(idUrl, empresaID)
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
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	idToken, err := h.converter.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	idUrl, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID do usuário deve ser um número"})
		return
	}

	if idUrl != idToken {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para editar este usuário"})
		return
	}

	var dadosParaAtualizar map[string]interface{}
	if err := c.ShouldBindJSON(&dadosParaAtualizar); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corpo da requisição (JSON) inválido"})
		return
	}

	err = h.service.Update(idUrl, empresaID, dadosParaAtualizar)
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
	_, err := h.empresaService.GetEmpresaByIDSer(request.EmpresaID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A empresa especificada não existe."})
		return
	}

	_, err = h.cargoService.FindByID(request.CargoID, request.EmpresaID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O cargo especificado não existe ou não pertence a esta empresa."})
		return
	}
	usuario := model.Usuario{
		Nome:      request.Nome,
		Email:     request.Email,
		Senha:     request.Senha,
		EmpresaID: request.EmpresaID,
		CargoID:   request.CargoID,
	}

	err = h.service.CriarUsuario(&usuario) // Este método agora será mais simples!
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, usuario)
}

func (h *UsuarioHandler) GetMeuPerfil(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := h.converter.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	usuario, err := h.service.FindByID(id, empresaID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	c.JSON(http.StatusOK, usuario)
}
