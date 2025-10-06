package auth

import (
	"errors"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

// PermissionMiddleware verifica se o cargo de um utilizador tem uma permissão específica.
func PermissionMiddleware(usuarioService usuario.UsuarioService, funcoesService funcoes.FuncoesInterface, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Obter os IDs do token (quem está a fazer o pedido?)
		empresaID, err := funcoesService.GetUintIDFromContext(c, "empresaID")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Falha ao verificar permissões: empresaID inválido"})
			return
		}
		userID, err := funcoesService.GetUintIDFromContext(c, "userID")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Falha ao verificar permissões: userID inválido"})
			return
		}

		user, err := usuarioService.FindByID(userID, empresaID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Acesso negado: utilizador não encontrado."})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar permissões."})
			return
		}

		if user.Contrato.ID == 0 || user.Contrato.Cargo.ID == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Acesso negado. O seu utilizador não tem um contrato ou cargo válido."})
			return
		}

		hasPermission := false
		for _, p := range user.Contrato.Cargo.Permissoes {
			if p.Nome == requiredPermission {
				hasPermission = true
				break 
			}
		}
		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Acesso negado. O seu cargo não tem permissão para executar esta ação."})
			return
		}

		// Se o utilizador tem a permissão, deixamo-lo continuar para o handler final.
		c.Next()
	}
}