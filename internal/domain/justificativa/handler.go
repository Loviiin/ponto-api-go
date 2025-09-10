// internal/domain/justificativa/handler.go
package justificativa

import (
	"net/http"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   Service
	converter funcoes.FuncoesInterface
}

func NewHandler(s Service, f funcoes.FuncoesInterface) *Handler {
	return &Handler{service: s, converter: f}
}

type solicitarAjusteRequest struct {
	DataOcorrencia time.Time `json:"data_ocorrencia" binding:"required" example:"2025-09-10T09:00:00Z"`
	Tipo           string    `json:"tipo" binding:"required" example:"ENTRADA_ESQUECIDA"`
	Descricao      string    `json:"descricao" binding:"required" example:"Esqueci de bater o ponto na entrada."`
}

type aprovarReprovarRequest struct {
	Aprovado         bool   `json:"aprovado"`
	MotivoReprovacao string `json:"motivo_reprovacao,omitempty"`
}

// @Summary      Solicita um ajuste de ponto
// @Description  Um funcionário cria uma solicitação para adicionar ou corrigir um registro de ponto.
// @Tags         Justificativas
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        solicitacao  body      solicitarAjusteRequest  true  "Dados da solicitação de ajuste"
// @Success      201          {object}  model.Justificativa
// @Failure      400          {object}  map[string]string
// @Failure      500          {object}  map[string]string "Falha interna no servidor"
// @Router       /justificativas [post]
func (h *Handler) SolicitarAjuste(c *gin.Context) {
	userID, _ := h.converter.GetUintIDFromContext(c, "userID")
	empresaID, _ := h.converter.GetUintIDFromContext(c, "empresaID")

	var req solicitarAjusteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	justificativa := model.Justificativa{
		UsuarioID:      userID,
		EmpresaID:      empresaID,
		DataOcorrencia: req.DataOcorrencia,
		Tipo:           req.Tipo,
		Descricao:      req.Descricao,
	}

	if err := h.service.SolicitarAjuste(&justificativa); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar solicitação"})
		return
	}

	c.JSON(http.StatusCreated, justificativa)
}

// @Summary      (Admin) Lista justificativas pendentes
// @Description  Retorna uma lista de todas as solicitações de ajuste de ponto que estão pendentes de aprovação. Requer permissão.
// @Tags         Justificativas
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Justificativa
// @Failure      500  {object}  map[string]string
// @Router       /justificativas/pendentes [get]
func (h *Handler) ListarPendentes(c *gin.Context) {
	empresaID, _ := h.converter.GetUintIDFromContext(c, "empresaID")

	pendentes, err := h.service.ListarPendentes(empresaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar solicitações pendentes"})
		return
	}
	c.JSON(http.StatusOK, pendentes)
}

// @Summary      (Admin) Aprova ou reprova uma justificativa
// @Description  Processa uma solicitação de ajuste, aprovando-a (o que cria o registro de ponto) ou reprovando-a. Requer permissão.
// @Tags         Justificativas
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      int                       true  "ID da Justificativa"
// @Param        acao   body      aprovarReprovarRequest    true  "Ação de Aprovação/Reprovação"
// @Success      200    {object}  model.Justificativa
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Router       /justificativas/{id}/processar [post]
func (h *Handler) AprovarReprovar(c *gin.Context) {
	aprovadorID, _ := h.converter.GetUintIDFromContext(c, "userID")
	empresaID, _ := h.converter.GetUintIDFromContext(c, "empresaID")

	justificativaID, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da justificativa inválido"})
		return
	}

	var req aprovarReprovarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	justificativa, err := h.service.AprovarReprovar(justificativaID, empresaID, aprovadorID, req.Aprovado, req.MotivoReprovacao)
	if err != nil {
		if err.Error() == "justificativa não encontrada" || err.Error() == "esta solicitação já foi processada" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao processar solicitação: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, justificativa)
}
