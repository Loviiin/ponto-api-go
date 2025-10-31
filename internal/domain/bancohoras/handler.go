package bancohoras

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/Loviiin/ponto-api-go/pkg/permissions"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service        BancoHorasService
	usuarioService usuario.UsuarioService
	converter      funcoes.FuncoesInterface
}

func NewBancoHorasHandler(s BancoHorasService, u usuario.UsuarioService, f funcoes.FuncoesInterface) *Handler {
	return &Handler{
		service:        s,
		usuarioService: u,
		converter:      f,
	}
}

// GetDashboard retorna saldo total e histórico do banco de horas do usuário autenticado
// @Summary      Meu Banco de Horas - Dashboard
// @Description  Retorna, em uma única chamada, o saldo total do banco de horas e o histórico de lançamentos já calculados para o usuário autenticado. Permite filtrar o histórico por data_inicio e data_fim.
// @Tags         Banco de Horas
// @Produce      json
// @Security     BearerAuth
// @Param        data_inicio  query     string  false  "Data inicial do histórico (formato: YYYY-MM-DD)"  example("2025-10-01")
// @Param        data_fim     query     string  false  "Data final do histórico (formato: YYYY-MM-DD)"    example("2025-10-31")
// @Success      200  {object}  bancohoras.DashboardResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /bancohoras/dashboard/me [get]
func (h *Handler) GetDashboard(c *gin.Context) {
	userID, err := h.converter.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "empresa não encontrada no token"})
		return
	}

	// Carrega timezone do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local
	}

	// Parâmetros opcionais de filtro
	var dataInicio, dataFim *time.Time

	dataInicioStr := c.Query("data_inicio")
	if dataInicioStr != "" {
		di, err := time.ParseInLocation("2006-01-02", dataInicioStr, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data_inicio inválido. Use YYYY-MM-DD."})
			return
		}
		dataInicio = &di
	}

	dataFimStr := c.Query("data_fim")
	if dataFimStr != "" {
		df, err := time.ParseInLocation("2006-01-02", dataFimStr, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data_fim inválido. Use YYYY-MM-DD."})
			return
		}
		dataFim = &df
	}

	resp, err := h.service.GetDashboardForUsuario(userID, empresaID, dataInicio, dataFim)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetSaldoUsuario retorna o saldo atual do banco de horas de um usuário específico
// @Summary      Saldo de Banco de Horas de um Funcionário
// @Description  Retorna o saldo atual do banco de horas de um usuário específico. Requer permissão 'VER_SALDO_FUNCIONARIOS'.
// @Tags         Banco de Horas
// @Produce      json
// @Security     BearerAuth
// @Param        userId   path      int  true  "ID do Usuário"
// @Success      200  {object}  map[string]int
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /bancohoras/dashboard/{userId} [get]
func (h *Handler) GetSaldoUsuario(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "empresa não encontrada no token"})
		return
	}

	userID, err := h.converter.StrParaUint(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do usuário inválido"})
		return
	}

	// Busca o saldo usando o service (com cache)
	saldo, err := h.service.GetSaldoAtualUsuario(userID, empresaID)
	if err != nil {
		if err.Error() == "usuário não possui contrato ativo" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"saldo_atual_minutos": saldo,
	})
}

// @Summary      Consulta saldo de horas do dia
// @Description  Retorna o saldo de horas (positivo ou negativo) de um usuário para um dia específico. Requer permissão 'VER_SALDO_FUNCIONARIOS' se o ID consultado não for o do próprio usuário.
// @Tags         Banco de Horas
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int     true  "ID do Usuário"
// @Param        dia  query     string  true  "Dia para consulta (formato: AAAA-MM-DD)"  example("2025-08-26")
// @Success      200  {object}  map[string]int
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /bancohoras/saldo/usuario/{id} [get]
func (h *Handler) GetSaldoDoDia(c *gin.Context) {
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

	diaString := c.Query("dia")

	// Carrega o timezone do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local // fallback
	}

	diaTime, err := time.ParseInLocation("2006-01-02", diaString, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido. Use AAAA-MM-DD."})
		return
	}

	idDoRequisitante, err := h.converter.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if idDoRequisitante != id {
		requisitante, err := h.usuarioService.FindByID(idDoRequisitante, empresaID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado."})
			return
		}

		temPermissao := false
		for _, permissao := range requisitante.Contrato.Cargo.Permissoes {
			if permissao.Nome == permissions.VER_SALDO_FUNCIONARIOS {
				temPermissao = true
				break
			}
		}

		if !temPermissao {
			c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para ver o saldo de outros funcionários."})
			return
		}
	}

	saldoEmMinutos, err := h.service.CalcularSaldoParaUsuario(id, empresaID, diaTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao calcular o saldo."})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"saldo_em_minutos": saldoEmMinutos,
	})
}

// @Summary      Realiza o fechamento manual de um dia
// @Description  Calcula o saldo de horas de um dia para um usuário e o adiciona ao saldo total do banco de horas. Requer permissão 'EDITAR_SALDO_FUNCIONARIOS'.
// @Tags         Banco de Horas
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int     true  "ID do Usuário"
// @Param        dia  query     string  true  "Dia para fechar (formato: AAAA-MM-DD)"  example("2025-08-26")
// @Success      200  {object}  model.Contrato
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /bancohoras/fechamento/usuario/{id} [post]
func (h *Handler) FecharDia(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID da empresa inválido no token"})
		return
	}
	idUsuarioAlvo, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID do usuário na URL deve ser um número"})
		return
	}
	diaString := c.Query("dia")
	if diaString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O parâmetro 'dia' é obrigatório. Use o formato AAAA-MM-DD."})
		return
	}

	// Carrega o timezone do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local // fallback
	}

	diaTime, err := time.ParseInLocation("2006-01-02", diaString, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido. Use AAAA-MM-DD."})
		return
	}

	usuarioAtualizado, err := h.service.FecharDiaParaUsuario(idUsuarioAlvo, empresaID, diaTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao processar o fechamento do dia: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, usuarioAtualizado)
}

func (h *Handler) ExecutarFechamentoDiario(c *gin.Context) {
	schedulerSecret := os.Getenv("SCHEDULER_SECRET")
	if schedulerSecret == "" {
		log.Println("ERRO CRÍTICO DE SEGURANÇA: A variável SCHEDULER_SECRET não está configurada.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Serviço mal configurado"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	splitToken := strings.Split(authHeader, " ")
	if len(splitToken) != 2 || splitToken[0] != "Bearer" || splitToken[1] != schedulerSecret {
		log.Printf("AVISO DE SEGURANÇA: Tentativa de acesso não autorizada ao endpoint de fechamento. IP: %s", c.ClientIP())
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Acesso não permitido"})
		return
	}

	log.Println("Iniciando tarefa agendada via HTTP: Fechamento diário do banco de horas...")
	diaAnterior := time.Now().AddDate(0, 0, -1)

	usuarios, err := h.usuarioService.FindAll()
	if err != nil {
		log.Printf("SCHEDULER_HTTP: Erro ao buscar usuários: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar usuários"})
		return
	}

	log.Printf("SCHEDULER_HTTP: Encontrados %d usuários para processar.", len(usuarios))

	for _, usr := range usuarios {
		_, err := h.service.FecharDiaParaUsuario(usr.ID, usr.Contrato.EmpresaID, diaAnterior)
		if err != nil {
			log.Printf("SCHEDULER_HTTP: Erro ao fechar o dia para o usuário ID %d: %v", usr.ID, err)
		} else {
			log.Printf("SCHEDULER_HTTP: Fechamento do dia %s concluído para o usuário ID %d.", diaAnterior.Format("2006-01-02"), usr.ID)
		}
	}

	log.Println("Tarefa agendada via HTTP: Fechamento diário concluído.")
	c.JSON(http.StatusOK, gin.H{"status": "Fechamento diário concluído com sucesso"})
}
