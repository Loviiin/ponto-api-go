package empresa

import (
	"errors"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type EmpresaHandler struct {
	service   EmpresaService
	converter funcoes.FuncoesInterface
	usuario   usuario.UsuarioService
}

func NewEmpresaHandler(s EmpresaService, f funcoes.FuncoesInterface, u usuario.UsuarioService) *EmpresaHandler {
	return &EmpresaHandler{
		service:   s,
		converter: f,
		usuario:   u,
	}
}

func (h *EmpresaHandler) CriarEmpresaHandler(c *gin.Context) {
	type criaEmpresaRequest struct {
		Nome               string  `json:"nome" binding:"required"`
		SedeLatitude       float64 `json:"sedeLatitude" binding:"required"`
		SedeLongitude      float64 `json:"sedeLongitude" binding:"required"`
		RaioGeofenceMetros float64 `json:"raioGeofenceMetros" binding:"required"`
	}

	var request criaEmpresaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	empresa := model.Empresa{
		Nome:               request.Nome,
		SedeLatitude:       request.SedeLatitude,
		SedeLongitude:      request.SedeLongitude,
		RaioGeofenceMetros: request.RaioGeofenceMetros,
	}

	err := h.service.CreateEmpresa(&empresa)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, empresa)
}

func (h *EmpresaHandler) GetAllEmpresasHandler(c *gin.Context) {

	empresas, err := h.service.GetAllEmpresasSer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar empresas"})
		return
	}
	c.JSON(http.StatusOK, empresas)

}

func (h *EmpresaHandler) GetEmpresaByIDHandler(c *gin.Context) {
	idEmpresaStr := c.Param("id")

	id, err := h.converter.StrParaUint(idEmpresaStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID da empresa deve ser um número válido"})
		return
	}

	empresa, err := h.service.GetEmpresaByIDSer(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Empresa não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar empresa"})
		return
	}

	c.JSON(http.StatusOK, empresa)
}

func (h *EmpresaHandler) UpdateEmpresaHandler(c *gin.Context) {
	idEmpresaURL := c.Param("id")

	idEmpresaToken, _ := c.Get("empresaID")
	idUsuarioToken, _ := c.Get("userID")

	if idEmpresaURL != idEmpresaToken.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado. Você não tem permissão para editar esta empresa."})
		return
	}

	idEmpresa, err := h.converter.StrParaUint(idEmpresaURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID da empresa na URL é inválido."})
		return
	}
	idUsuario, _ := h.converter.StrParaUint(idUsuarioToken.(string))

	usuario, err := h.usuario.FindByID(idUsuario, idEmpresa)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário de autenticação não encontrado."})
		return
	}

	// TODO: Refatorar esta lógica para usar o motor de RBAC (Issue #29)
	if usuario.Cargo.Nome == "ADMIN" { // <<<<<< SUA LÓGICA DE CARGO AQUI
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado. Seu cargo não tem permissão para executar esta ação."})
		return
	}

	var dadosParaAtualizar map[string]interface{}
	if err := c.ShouldBindJSON(&dadosParaAtualizar); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corpo da requisição (JSON) inválido"})
		return
	}

	if err := h.service.UpdateEmpresaSer(idEmpresa, dadosParaAtualizar); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar a empresa"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *EmpresaHandler) DeleteEmpresaHandler(c *gin.Context) {

	idEmpresaURL := c.Param("id")

	idEmpresaToken, _ := c.Get("empresaID")
	idUsuarioToken, _ := c.Get("userID")

	if idEmpresaURL != idEmpresaToken.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado. Você não tem permissão para deletar esta empresa."})
		return
	}

	idEmpresa, err := h.converter.StrParaUint(idEmpresaURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID da empresa na URL é inválido."})
		return
	}
	idUsuario, _ := h.converter.StrParaUint(idUsuarioToken.(string))

	usuario, err := h.usuario.FindByID(idUsuario, idEmpresa)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário de autenticação não encontrado."})
		return
	}

	// TODO: Refatorar esta lógica para usar o motor de RBAC (Issue #29)
	if usuario.Cargo.Nome != "ADMIN" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado. Seu cargo não tem permissão para executar esta ação."})
		return
	}

	err = h.service.DeleteEmpresaSer(idEmpresa)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Empresa não encontrada para deletar."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao deletar a empresa."})
		return
	}

	c.Status(http.StatusNoContent)
}
