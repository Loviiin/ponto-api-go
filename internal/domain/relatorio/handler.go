package relatorio

import (
	"net/http"
	"time"

	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
	conv    funcoes.FuncoesInterface
}

func NewHandler(service Service, conv funcoes.FuncoesInterface) *Handler {
	return &Handler{service: service, conv: conv}
}

// GetEspelhoMe retorna o espelho de ponto (resumo consolidado) do usuário autenticado.
// @Summary      Obtém espelho de ponto do usuário logado
// @Description  Retorna o espelho de ponto (cálculo consolidado) para o intervalo informado.
// @Tags         Relatórios
// @Produce      json
// @Security     BearerAuth
// @Param        data_inicio  query     string  true  "Data inicial (AAAA-MM-DD)"
// @Param        data_fim     query     string  true  "Data final (AAAA-MM-DD)"
// @Success      200          {object}  GerarEspelhoResponse
// @Failure      400          {object}  map[string]string
// @Failure      401          {object}  map[string]string
// @Failure      500          {object}  map[string]string
// @Router       /relatorios/espelho/me [get]
func (h *Handler) GetEspelhoMe(c *gin.Context) {
	uid, err := h.conv.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "userID inválido"})
		return
	}
	empID, err := h.conv.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "empresaID inválido"})
		return
	}
	inicioStr := c.Query("data_inicio")
	fimStr := c.Query("data_fim")
	if inicioStr == "" || fimStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetros data_inicio e data_fim são obrigatórios"})
		return
	}
	inicio, err := time.Parse("2006-01-02", inicioStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_inicio inválida"})
		return
	}
	fim, err := time.Parse("2006-01-02", fimStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_fim inválida"})
		return
	}
	// Ajustar fim para incluir final do dia
	fim = fim.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	espelho, err := h.service.GerarEspelhoPonto(uid, empID, inicio, fim)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, espelho)
}

// GetEspelhoUsuario retorna o espelho de ponto de um usuário específico (admin).
// @Summary      (Admin) Obtém espelho de ponto de um usuário
// @Description  Retorna o espelho de ponto de um usuário da empresa para o intervalo informado. Requer permissão administrativa.
// @Tags         Relatórios
// @Produce      json
// @Security     BearerAuth
// @Param        id           path      int     true   "ID do Usuário"
// @Param        data_inicio  query     string  true   "Data inicial (AAAA-MM-DD)"
// @Param        data_fim     query     string  true   "Data final (AAAA-MM-DD)"
// @Success      200          {object}  GerarEspelhoResponse
// @Failure      400          {object}  map[string]string
// @Failure      401          {object}  map[string]string
// @Failure      500          {object}  map[string]string
// @Router       /relatorios/espelho/usuario/{id} [get]
func (h *Handler) GetEspelhoUsuario(c *gin.Context) {
	empID, err := h.conv.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "empresaID inválido"})
		return
	}
	uidParam, err := h.conv.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}
	inicioStr := c.Query("data_inicio")
	fimStr := c.Query("data_fim")
	if inicioStr == "" || fimStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetros data_inicio e data_fim são obrigatórios"})
		return
	}
	inicio, err := time.Parse("2006-01-02", inicioStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_inicio inválida"})
		return
	}
	fim, err := time.Parse("2006-01-02", fimStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_fim inválida"})
		return
	}
	fim = fim.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	espelho, err := h.service.GerarEspelhoPonto(uidParam, empID, inicio, fim)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, espelho)
}
