package ponto

import (
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type PontoHandler struct {
	service   PontoService
	converter funcoes.FuncoesInterface
}

func NewPontoHandler(service PontoService, converter funcoes.FuncoesInterface) *PontoHandler {
	return &PontoHandler{
		service:   service,
		converter: converter,
	}
}

type BaterPontoRequest struct {
	Latitude  float64 `json:"latitude"  example:"-15.799879"`
	Longitude float64 `json:"longitude" example:"-47.864162"`
}

type AjustePontoRequest struct {
	UsuarioID     uint      `json:"usuario_id" binding:"required"`
	Timestamp     time.Time `json:"timestamp" binding:"required"`
	Justificativa string    `json:"justificativa" binding:"required"`
}

// @Summary      Registra uma batida de ponto
// @Description  Registra um evento de ponto (entrada/saída) para o usuário logado.
// @Tags         Ponto
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        ponto  body      BaterPontoRequest  true  "Coordenadas da Batida de Ponto"
// @Success      201    {object}  model.RegistroPonto
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /pontos [post]
func (h *PontoHandler) BaterPonto(c *gin.Context) {
	valorIDToken, existe := c.Get("userID")
	if !existe {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID do usuário não encontrado no contexto"})
		return
	}
	idTokenString, ok := valorIDToken.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID do usuário no contexto está em formato inválido"})
		return
	}
	usuarioID, err := strconv.ParseUint(idTokenString, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID do usuário no token é inválido"})
		return
	}

	valorEmpresaID, existe := c.Get("empresaID")
	if !existe {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID da empresa não encontrado no contexto"})
		return
	}
	idEmpresaString, ok := valorEmpresaID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID da empresa no contexto está em formato inválido"})
		return
	}
	empresaID, err := strconv.ParseUint(idEmpresaString, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID da empresa no token é inválido"})
		return
	}
	var requisicao BaterPontoRequest
	if err := c.ShouldBindJSON(&requisicao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corpo da requisição (JSON) inválido"})
		return
	}

	pontoRegistrado, err := h.service.BaterPonto(uint(usuarioID), uint(empresaID), requisicao.Latitude, requisicao.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao registrar o ponto"})
		return
	}

	c.JSON(http.StatusCreated, pontoRegistrado)
}

// --- INÍCIO DA NOVA FUNÇÃO ---

// @Summary      (Admin) Lista os registros de ponto de um usuário
// @Description  Retorna uma lista das batidas de ponto de um usuário específico para um determinado dia. Requer permissão 'VISUALIZAR_PONTO_FUNCIONARIOS'.
// @Tags         Ponto
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int     true   "ID do Usuário"
// @Param        dia  query     string  false  "Dia para consulta (formato: AAAA-MM-DD)"  example("2025-08-26")
// @Success      200  {array}   model.RegistroPonto
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /pontos/usuario/{id} [get]
func (h *PontoHandler) GetRegistosPorUsuarioID(c *gin.Context) {
	idUsuarioAlvo, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido."})
		return
	}
	diaQuery := c.Query("dia")
	var dia time.Time
	if diaQuery == "" {
		dia = time.Now()
	} else {
		dia, err = time.Parse("2006-01-02", diaQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido."})
			return
		}
	}
	registos, err := h.service.GetPontosDoDia(uint(idUsuarioAlvo), dia)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar registros."})
		return
	}
	c.JSON(http.StatusOK, registos)
}

// @Summary      Lista os registros de ponto do usuário
// @Description  Retorna uma lista das batidas de ponto do usuário logado para um dia específico. Se o dia não for fornecido, retorna os do dia atual.
// @Tags         Ponto
// @Produce      json
// @Security     BearerAuth
// @Param        dia  query     string  false  "Dia para consulta (formato: AAAA-MM-DD)"  example("2025-08-26")
// @Success      200  {array}   model.RegistroPonto
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /pontos/meus-registros [get]
func (h *PontoHandler) GetMeusRegistos(c *gin.Context) {
	valorIDToken, _ := c.Get("userID")
	idTokenString, _ := valorIDToken.(string)
	usuarioID, err := strconv.ParseUint(idTokenString, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID do utilizador no token é inválido"})
		return
	}

	diaQuery := c.Query("dia")
	var dia time.Time

	if diaQuery == "" {
		dia = time.Now()
	} else {
		dia, err = time.Parse("2006-01-02", diaQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido. Use AAAA-MM-DD."})
			return
		}
	}

	registos, err := h.service.GetPontosDoDia(uint(usuarioID), dia)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar os registos de ponto"})
		return
	}

	c.JSON(http.StatusOK, registos)
}

// @Summary      (Admin) Adiciona um registro de ponto manual
// @Description  Adiciona um novo registro de ponto para um funcionário com uma justificativa. Requer permissão 'AJUSTAR_PONTO_FUNCIONARIOS'.
// @Tags         Ponto
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        ajuste  body      AjustePontoRequest  true  "Dados do ajuste de ponto"
// @Success      201     {object}  model.RegistroPonto
// @Failure      400     {object}  map[string]string
// @Failure      403     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /pontos/ajuste [post]
func (h *PontoHandler) AjustarPonto(c *gin.Context) {
	idAdmin, err := h.converter.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID do admin inválido."})
		return
	}
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID da empresa inválido."})
		return
	}

	var req AjustePontoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corpo da requisição inválido: " + err.Error()})
		return
	}

	novoPonto, err := h.service.AjustarPonto(req.UsuarioID, empresaID, idAdmin, req.Timestamp, req.Justificativa)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao ajustar o ponto: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, novoPonto)
}
