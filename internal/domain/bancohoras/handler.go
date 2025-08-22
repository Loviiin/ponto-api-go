package bancohoras

import (
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/Loviiin/ponto-api-go/pkg/permissions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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

// GetSaldoDoDia é a função que vai lidar com a requisição da API.
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
	diaTime, err := time.Parse("2006-01-02", diaString)
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
		for _, permissao := range requisitante.Cargo.Permissoes {
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
	diaTime, err := time.Parse("2006-01-02", diaString)
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
