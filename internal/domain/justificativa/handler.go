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
	DataOcorrencia time.Time  `json:"data_ocorrencia" binding:"required" example:"2025-09-10T09:00:00Z"`
	Tipo           string     `json:"tipo" binding:"required,oneof=PONTO_FALTANTE CORRECAO_PONTO" example:"PONTO_FALTANTE"`
	Descricao      string     `json:"descricao" binding:"required,min=10" example:"Esqueci de bater o ponto na entrada."`
	PontoID        *uint      `json:"ponto_id,omitempty" example:"123"`                      // Obrigatório para CORRECAO_PONTO
	NovoHorario    *time.Time `json:"novo_horario,omitempty" example:"2025-09-10T08:00:00Z"` // Obrigatório para PONTO_FALTANTE
}

type solicitarCorrecaoRequest struct {
	PontoID      uint      `json:"ponto_id" binding:"required" example:"123"`
	NovaDataHora time.Time `json:"nova_data_hora" binding:"required" example:"2025-09-10T08:00:00Z"`
	Descricao    string    `json:"descricao" binding:"required" example:"Bati o ponto com atraso devido ao trânsito intenso."`
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

	// Validações customizadas
	if req.Tipo == "CORRECAO_PONTO" {
		if req.PontoID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CORRECAO_PONTO requer ponto_id"})
			return
		}
		if req.NovoHorario == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CORRECAO_PONTO requer novo_horario"})
			return
		}
	}

	if req.Tipo == "PONTO_FALTANTE" {
		if req.NovoHorario == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "PONTO_FALTANTE requer novo_horario"})
			return
		}
		if req.PontoID != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "PONTO_FALTANTE não deve ter ponto_id"})
			return
		}
	}

	justificativa := model.Justificativa{
		UsuarioID:      userID,
		EmpresaID:      empresaID,
		DataOcorrencia: req.DataOcorrencia,
		Tipo:           req.Tipo,
		Descricao:      req.Descricao,
		PontoID:        req.PontoID,
		NovoHorario:    req.NovoHorario,
	}

	if err := h.service.SolicitarAjuste(&justificativa); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, justificativa)
}

// @Summary      Solicita correção de ponto existente
// @Description  Funcionário solicita correção de um ponto que já foi batido (com horário incorreto/atrasado).
// @Tags         Justificativas
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        correcao  body      solicitarCorrecaoRequest  true  "Dados da solicitação de correção"
// @Success      201       {object}  map[string]string
// @Failure      400       {object}  map[string]string
// @Failure      404       {object}  map[string]string
// @Router       /justificativas/solicitar-correcao [post]
func (h *Handler) SolicitarCorrecaoPonto(c *gin.Context) {
	userID, _ := h.converter.GetUintIDFromContext(c, "userID")
	empresaID, _ := h.converter.GetUintIDFromContext(c, "empresaID")

	var req solicitarCorrecaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.SolicitarCorrecaoPonto(req.PontoID, req.NovaDataHora, req.Descricao, userID, empresaID)
	if err != nil {
		if err.Error() == "ponto não encontrado" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Solicitação de correção criada com sucesso"})
}

// @Summary      (Admin) Lista justificativas pendentes
// @Description  Retorna uma lista de todas as solicitações de ajuste de ponto que estão pendentes de aprovação. Requer permissão.
// @Tags         Justificativas
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Justificativa  "Exemplo"
// @Example 200 [{"id":55,"usuario_id":3,"empresa_id":1,"tipo":"ENTRADA_ESQUECIDA","descricao":"Esqueci de bater ao chegar","status":"PENDENTE"}]
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

// @Summary      Lista minhas justificativas
// @Description  Retorna todas as solicitações de ajuste de ponto do usuário autenticado (independente do status).
// @Tags         Justificativas
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Justificativa  "Lista de justificativas do usuário"
// @Failure      500  {object}  map[string]string
// @Router       /justificativas/minhas [get]
func (h *Handler) ListarMinhas(c *gin.Context) {
	userID, _ := h.converter.GetUintIDFromContext(c, "userID")
	empresaID, _ := h.converter.GetUintIDFromContext(c, "empresaID")

	minhas, err := h.service.ListarPorUsuario(userID, empresaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar suas justificativas"})
		return
	}
	c.JSON(http.StatusOK, minhas)
}

// @Summary      Cancela uma solicitação própria
// @Description  Permite que o usuário cancele sua própria solicitação de ajuste que ainda está pendente.
// @Tags         Justificativas
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID da Justificativa a ser cancelada"
// @Success      200  {object}  map[string]string  "Solicitação cancelada com sucesso"
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /justificativas/{id}/cancelar [delete]
func (h *Handler) CancelarSolicitacao(c *gin.Context) {
	userID, _ := h.converter.GetUintIDFromContext(c, "userID")
	empresaID, _ := h.converter.GetUintIDFromContext(c, "empresaID")

	justificativaID, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da justificativa inválido"})
		return
	}

	err = h.service.CancelarSolicitacao(justificativaID, empresaID, userID)
	if err != nil {
		errorMsg := err.Error()
		switch errorMsg {
		case "justificativa não encontrada":
			c.JSON(http.StatusNotFound, gin.H{"error": "Justificativa não encontrada"})
		case "você só pode cancelar suas próprias solicitações":
			c.JSON(http.StatusForbidden, gin.H{"error": errorMsg})
		case "apenas solicitações pendentes podem ser canceladas":
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao cancelar solicitação"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Solicitação cancelada com sucesso"})
}
