package ponto

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	apperrors "github.com/Loviiin/ponto-api-go/pkg/errors"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PontoHandler struct {
	service              PontoService
	justificativaService JustificativaService // Usa interface local para quebrar ciclo
	bancoHorasService    BancoHorasService    // Para invalidar cache quando há ajustes
	converter            funcoes.FuncoesInterface
}

// Interface local para quebrar ciclo de importação
type JustificativaService interface {
	SolicitarAjuste(j *model.Justificativa) error
}

// Interface local para invalidar cache de banco de horas
type BancoHorasService interface {
	InvalidarCacheDia(usuarioID uint, empresaID uint, dia time.Time)
	InvalidarCacheUsuario(usuarioID uint, empresaID uint)
}

// 3. Receber o serviço de justificativa no construtor
func NewPontoHandler(service PontoService, justificativaService JustificativaService, bancoHorasService BancoHorasService, converter funcoes.FuncoesInterface) *PontoHandler {
	return &PontoHandler{
		service:              service,
		justificativaService: justificativaService,
		bancoHorasService:    bancoHorasService,
		converter:            converter,
	}
}

type BaterPontoRequest struct {
	Latitude  float64 `json:"latitude"  example:"-15.799879"`
	Longitude float64 `json:"longitude" example:"-47.864162"`
}

type AjustePontoRequest struct {
	UsuarioID     uint      `json:"usuario_id" binding:"required" example:"2"`
	Timestamp     time.Time `json:"timestamp" binding:"required" example:"2025-09-10T09:00:00Z"`
	Justificativa string    `json:"justificativa" binding:"required" example:"Ajuste manual de entrada."`
}

type EditarPontoRequest struct {
	Timestamp     time.Time `json:"timestamp" binding:"required" example:"2025-09-10T18:05:00Z"`
	Justificativa string    `json:"justificativa" binding:"required" example:"Correção do horário de saída."`
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
		err := apperrors.NewInternal(errors.New("ID do usuário não encontrado no contexto"))
		c.JSON(err.Code, err)
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

	pontoRegistrado, err := h.service.BaterPonto(c.Request.Context(), uint(usuarioID), uint(empresaID), requisicao.Latitude, requisicao.Longitude)
	if err != nil {
		slog.Error("Falha ao registrar ponto", "error", err, "user_id", usuarioID)
		appErr := apperrors.NewInternal(err)
		c.JSON(appErr.Code, appErr)
		return
	}

	c.JSON(http.StatusCreated, pontoRegistrado)
}

// @Summary      (Admin) Lista os registros de ponto de um usuário
// @Description  Retorna uma lista das batidas de ponto de um usuário específico para um determinado dia. Requer permissão 'VISUALIZAR_PONTO_FUNCIONARIOS'.
// @Tags         Ponto
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int     true   "ID do Usuário"
// @Param        dia  query     string  false  "Dia para consulta (formato: AAAA-MM-DD)"  example("2025-08-26")
// @Success      200  {array}   model.RegistroPonto  "Exemplo"
// @Example 200 [{"id":100,"timestamp":"2025-10-07T08:00:00Z","latitude":-15.79,"longitude":-47.86,"metodo":"Presencial"}]
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

	// Carrega o timezone do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local // fallback
	}

	if diaQuery == "" {
		dia = time.Now().In(loc)
	} else {
		// Parse a data no timezone do Brasil, não em UTC
		dia, err = time.ParseInLocation("2006-01-02", diaQuery, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido."})
			return
		}
	}
	registos, err := h.service.GetPontosDoDia(c.Request.Context(), uint(idUsuarioAlvo), dia)
	if err != nil {
		slog.Error("Falha ao buscar registros", "error", err, "user_id", idUsuarioAlvo)
		appErr := apperrors.NewInternal(err)
		c.JSON(appErr.Code, appErr)
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
// @Success      200  {array}   model.RegistroPonto  "Exemplo"
// @Example 200 [{"id":101,"timestamp":"2025-10-07T08:02:10Z","latitude":-15.79,"longitude":-47.86,"metodo":"Remoto"}]
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

	// Carrega o timezone do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local // fallback
	}

	if diaQuery == "" {
		dia = time.Now().In(loc)
	} else {
		// Parse a data no timezone do Brasil, não em UTC
		dia, err = time.ParseInLocation("2006-01-02", diaQuery, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido. Use AAAA-MM-DD."})
			return
		}
	}

	registos, err := h.service.GetPontosDoDia(c.Request.Context(), uint(usuarioID), dia)
	if err != nil {
		slog.Error("Falha ao buscar meus registros", "error", err, "user_id", usuarioID)
		appErr := apperrors.NewInternal(err)
		c.JSON(appErr.Code, appErr)
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

	// LÓGICA DE ORQUESTRAÇÃO:
	// 1. Criar a justificativa primeiro, usando o serviço de justificativa.
	justificativa := &model.Justificativa{
		UsuarioID:      req.UsuarioID,
		EmpresaID:      empresaID,
		AprovadorID:    &idAdmin,
		DataOcorrencia: req.Timestamp,
		Tipo:           "AJUSTE_ADMIN",
		Descricao:      req.Justificativa,
		Status:         "APROVADO", // Um ajuste feito por admin já nasce aprovado.
	}
	if err := h.justificativaService.SolicitarAjuste(justificativa); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar registro de justificativa: " + err.Error()})
		return
	}

	// 2. Chamar o serviço de ponto, passando o ID da justificativa que acabamos de criar.
	novoPonto, err := h.service.AjustarPonto(c.Request.Context(), req.UsuarioID, empresaID, idAdmin, req.Timestamp, &justificativa.ID)
	if err != nil {
		slog.Error("Falha ao ajustar ponto", "error", err, "admin_id", idAdmin, "target_user_id", req.UsuarioID)
		appErr := apperrors.NewInternal(err)
		c.JSON(appErr.Code, appErr)
		return
	}

	// 3. Invalidar cache do banco de horas para o dia ajustado
	if h.bancoHorasService != nil {
		h.bancoHorasService.InvalidarCacheDia(req.UsuarioID, empresaID, req.Timestamp)
		// Também invalida o dashboard e saldo geral (podem ter mudado)
		h.bancoHorasService.InvalidarCacheUsuario(req.UsuarioID, empresaID)
	}

	c.JSON(http.StatusCreated, novoPonto)
}

// @Summary      (Admin) Edita um registro de ponto existente
// @Description  Altera o timestamp de um registro de ponto existente, criando uma justificativa para a auditoria. Requer permissão 'AJUSTAR_PONTO_FUNCIONARIOS'.
// @Tags         Ponto
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        pontoId  path      int                 true  "ID do Registro de Ponto a ser editado"
// @Param        edicao   body      EditarPontoRequest  true  "Novos dados para o registro de ponto"
// @Success      200      {object}  model.RegistroPonto
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /pontos/{pontoId} [put]
func (h *PontoHandler) EditarPonto(c *gin.Context) {
	idAdmin, err := h.converter.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID do administrador inválido no token."})
		return
	}
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID da empresa inválido no token."})
		return
	}
	pontoID, err := h.converter.StrParaUint(c.Param("pontoId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID do registro de ponto na URL é inválido."})
		return
	}

	var req EditarPontoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corpo da requisição inválido: " + err.Error()})
		return
	}

	pontoOriginal, err := h.service.FindPontoByID(c.Request.Context(), pontoID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Registro de ponto não encontrado ou não pertence a esta empresa."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar o registro de ponto original: " + err.Error()})
		return
	}

	if pontoOriginal == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ocorreu um erro inesperado ao buscar o registro de ponto."})
		return
	}

	justificativa := &model.Justificativa{
		UsuarioID:      pontoOriginal.UsuarioID,
		EmpresaID:      empresaID,
		AprovadorID:    &idAdmin,
		DataOcorrencia: req.Timestamp,
		Tipo:           "EDICAO_ADMIN",
		Descricao:      req.Justificativa,
		Status:         "APROVADO",
	}
	if err := h.justificativaService.SolicitarAjuste(justificativa); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar registro de justificativa para edição: " + err.Error()})
		return
	}

	pontoAtualizado, err := h.service.EditarPonto(c.Request.Context(), pontoID, empresaID, req.Timestamp, &justificativa.ID)
	if err != nil {
		slog.Error("Falha ao editar ponto", "error", err, "ponto_id", pontoID)
		appErr := apperrors.NewInternal(err)
		c.JSON(appErr.Code, appErr)
		return
	}

	// Invalidar cache do banco de horas para os dias afetados (original e novo)
	if h.bancoHorasService != nil {
		// Invalida o dia original do ponto
		h.bancoHorasService.InvalidarCacheDia(pontoOriginal.UsuarioID, empresaID, pontoOriginal.Timestamp)
		// Invalida o novo dia (caso seja diferente)
		h.bancoHorasService.InvalidarCacheDia(pontoOriginal.UsuarioID, empresaID, req.Timestamp)
		// Invalida dashboard e saldo geral
		h.bancoHorasService.InvalidarCacheUsuario(pontoOriginal.UsuarioID, empresaID)
	}

	c.JSON(http.StatusOK, pontoAtualizado)
}

// ExportarRelatorio gera um relatório de registros de ponto em CSV ou PDF.
// Pode ser usado tanto pelo usuário para seus próprios registros quanto por administradores para usuários específicos.
// @Summary      Exporta relatório de registros de ponto
// @Description  Gera um arquivo (CSV ou PDF) contendo os registros de ponto no intervalo especificado.
// @Tags         Ponto
// @Produce      application/json
// @Security     BearerAuth
// @Param        formato      query     string  true   "Formato do arquivo (csv ou pdf)"  Enums(csv,pdf)
// @Param        data_inicio  query     string  true   "Data inicial (AAAA-MM-DD)"
// @Param        data_fim     query     string  true   "Data final (AAAA-MM-DD)"
// @Param        id           path      int     false  "(Admin) ID do usuário para exportar (usar rota /pontos/usuario/{id}/export)"
// @Success      200          "Arquivo gerado"
// @Failure      400          {object}  map[string]string
// @Failure      401          {object}  map[string]string
// @Failure      500          {object}  map[string]string
// @Router       /pontos/meus-registros/export [get]
// @Router       /pontos/usuario/{id}/export [get]
func (h *PontoHandler) ExportarRelatorio(c *gin.Context) {
	formato := c.Query("formato")
	dataInicioStr := c.Query("data_inicio")
	dataFimStr := c.Query("data_fim")

	if formato == "" || dataInicioStr == "" || dataFimStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetros obrigatórios: data_inicio, data_fim, formato"})
		return
	}

	// Carrega o timezone do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local // fallback
	}

	inicio, err := time.ParseInLocation("2006-01-02", dataInicioStr, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_inicio inválida. Use AAAA-MM-DD"})
		return
	}
	fim, err := time.ParseInLocation("2006-01-02", dataFimStr, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_fim inválida. Use AAAA-MM-DD"})
		return
	}
	fim = fim.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var userID uint
	if idParam := c.Param("id"); idParam != "" { // rota admin
		idParsed, err := h.converter.StrParaUint(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido na rota"})
			return
		}
		userID = idParsed
	} else {
		uid, err := h.converter.GetUintIDFromContext(c, "userID")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ID do usuário não encontrado no token"})
			return
		}
		userID = uid
	}

	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID da empresa não encontrado no token"})
		return
	}

	bytesArquivo, contentType, filename, err := h.service.GerarRelatorio(c.Request.Context(), userID, empresaID, inicio, fim, formato)
	if err != nil {
		if errors.Is(err, ErrFormatoInvalido) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao gerar relatório: " + err.Error()})
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, contentType, bytesArquivo)
}
