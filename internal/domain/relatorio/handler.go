package relatorio

import (
	"net/http"
	"strconv"
	"strings"
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
// @Description  Retorna o espelho de ponto (cálculo consolidado) para o intervalo informado. O saldo diário é calculado pelo serviço de Banco de Horas para garantir consistência.
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

	// Carrega o timezone do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local // fallback
	}

	inicio, err := time.ParseInLocation("2006-01-02", inicioStr, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_inicio inválida"})
		return
	}
	fim, err := time.ParseInLocation("2006-01-02", fimStr, loc)
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
// @Description  Retorna o espelho de ponto de um usuário da empresa para o intervalo informado. O saldo diário é calculado pelo serviço de Banco de Horas para garantir consistência. Requer permissão administrativa.
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

	// Carrega o timezone do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local // fallback
	}

	inicio, err := time.ParseInLocation("2006-01-02", inicioStr, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_inicio inválida"})
		return
	}
	fim, err := time.ParseInLocation("2006-01-02", fimStr, loc)
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

// GerarRelatorioGeral gera um relatório geral de ponto para um período específico
// @Summary      Gera relatório geral de ponto
// @Description  Retorna um relatório detalhado de ponto para um período, incluindo horas trabalhadas, extras, faltantes e banco de horas. A visualização de dados de outros usuários respeita a hierarquia de cargos: usuários só podem visualizar dados de funcionários com nível hierárquico menor. SuperAdmin e usuários com a permissão VISUALIZAR_RELATORIOS_GERAIS podem acessar todos os dados.
// @Tags         Relatórios
// @Produce      json
// @Security     BearerAuth
// @Param        data_inicio  query     string  true   "Data inicial (YYYY-MM-DD)"
// @Param        data_fim     query     string  true   "Data final (YYYY-MM-DD)"
// @Param        usuario_id   query     int     false  "ID do usuário (opcional, se omitido retorna todos os usuários ativos)"
// @Success      200          {object}  RelatorioGeralDTO
// @Failure      400          {object}  map[string]string
// @Failure      401          {object}  map[string]string
// @Failure      403          {object}  map[string]string  "Sem permissão ou tentativa de acessar usuário com nível hierárquico igual ou superior"
// @Failure      500          {object}  map[string]string
// @Router       /relatorios/geral [get]
func (h *Handler) GerarRelatorioGeral(c *gin.Context) {
	// Obter informações do usuário logado do contexto JWT
	userIDLogado, err := h.conv.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "userID inválido"})
		return
	}
	empresaID, err := h.conv.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "empresaID inválido"})
		return
	}

	// Validar parâmetros de data
	inicioStr := c.Query("data_inicio")
	fimStr := c.Query("data_fim")
	if inicioStr == "" || fimStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetros data_inicio e data_fim são obrigatórios"})
		return
	}

	// Carrega o timezone do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local // fallback
	}

	dataInicio, err := time.ParseInLocation("2006-01-02", inicioStr, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_inicio inválida (use formato YYYY-MM-DD)"})
		return
	}

	dataFim, err := time.ParseInLocation("2006-01-02", fimStr, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_fim inválida (use formato YYYY-MM-DD)"})
		return
	}

	// Ajustar data_fim para incluir o dia completo
	dataFim = dataFim.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// Processar parâmetro usuario_id (opcional)
	var usuarioID *uint
	usuarioIDStr := c.Query("usuario_id")
	if usuarioIDStr != "" {
		uid, err := h.conv.StrParaUint(usuarioIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "usuario_id inválido"})
			return
		}
		usuarioID = &uid
	}

	// Chamar o serviço com validação de hierarquia
	relatorio, err := h.service.GerarRelatorioGeral(dataInicio, dataFim, usuarioID, empresaID, userIDLogado)
	if err != nil {
		// Se o erro for de hierarquia, retornar 403
		if strings.Contains(err.Error(), "hierarquia") || strings.Contains(err.Error(), "permissão") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, relatorio)
}

// ExportarRelatorioGeral exporta um relatório geral de ponto em CSV ou PDF
// @Summary      Exporta relatório geral de ponto
// @Description  Gera um arquivo (CSV ou PDF) com o relatório geral de ponto para um período específico. Suporta paginação para relatórios de múltiplos usuários. A exportação de dados de outros usuários respeita a hierarquia de cargos: usuários só podem exportar dados de funcionários com nível hierárquico menor. SuperAdmin e usuários com a permissão VISUALIZAR_RELATORIOS_GERAIS podem exportar todos os dados.
// @Tags         Relatórios
// @Produce      application/json
// @Security     BearerAuth
// @Param        formato      query     string  true   "Formato do arquivo (csv ou pdf)"  Enums(csv,pdf)
// @Param        data_inicio  query     string  true   "Data inicial (YYYY-MM-DD)"
// @Param        data_fim     query     string  true   "Data final (YYYY-MM-DD)"
// @Param        usuario_id   query     int     false  "ID do usuário (opcional, se omitido exporta todos)"
// @Param        limite       query     int     false  "Limite de usuários por página (padrão: 50)"
// @Param        offset       query     int     false  "Deslocamento para paginação (padrão: 0)"
// @Success      200          "Arquivo gerado"
// @Failure      400          {object}  map[string]string
// @Failure      401          {object}  map[string]string
// @Failure      403          {object}  map[string]string  "Sem permissão ou tentativa de acessar usuário com nível hierárquico igual ou superior"
// @Failure      500          {object}  map[string]string
// @Router       /relatorios/geral/export [get]
func (h *Handler) ExportarRelatorioGeral(c *gin.Context) {
	// Obter informações do usuário logado do contexto JWT
	userIDLogado, err := h.conv.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "userID inválido"})
		return
	}
	empresaID, err := h.conv.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "empresaID inválido"})
		return
	}

	// Validar parâmetros obrigatórios
	formato := c.Query("formato")
	inicioStr := c.Query("data_inicio")
	fimStr := c.Query("data_fim")

	if formato == "" || inicioStr == "" || fimStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetros obrigatórios: formato, data_inicio, data_fim"})
		return
	}

	// Carrega o timezone do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local // fallback
	}

	// Validar datas
	dataInicio, err := time.ParseInLocation("2006-01-02", inicioStr, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_inicio inválida (use formato YYYY-MM-DD)"})
		return
	}

	dataFim, err := time.ParseInLocation("2006-01-02", fimStr, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_fim inválida (use formato YYYY-MM-DD)"})
		return
	}

	// Ajustar data_fim para incluir o dia completo
	dataFim = dataFim.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// Processar parâmetro usuario_id (opcional)
	var usuarioID *uint
	usuarioIDStr := c.Query("usuario_id")
	if usuarioIDStr != "" {
		uid, err := h.conv.StrParaUint(usuarioIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "usuario_id inválido"})
			return
		}
		usuarioID = &uid
	}

	// Parâmetros de paginação (apenas para relatórios de todos os usuários)
	limite := 50 // padrão
	offset := 0

	if limiteStr := c.Query("limite"); limiteStr != "" {
		if limiteVal, err := strconv.Atoi(limiteStr); err == nil && limiteVal > 0 {
			limite = limiteVal
			if limite > 200 { // máximo para evitar relatórios muito grandes
				limite = 200
			}
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offsetVal, err := strconv.Atoi(offsetStr); err == nil && offsetVal >= 0 {
			offset = offsetVal
		}
	}

	// Chamar o serviço de exportação com validação de hierarquia
	bytesArquivo, contentType, filename, err := h.service.ExportarRelatorioGeral(dataInicio, dataFim, usuarioID, empresaID, userIDLogado, formato, limite, offset)
	if err != nil {
		// Se o erro for de hierarquia, retornar 403
		if strings.Contains(err.Error(), "hierarquia") || strings.Contains(err.Error(), "permissão") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao gerar relatório: " + err.Error()})
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, contentType, bytesArquivo)
}
