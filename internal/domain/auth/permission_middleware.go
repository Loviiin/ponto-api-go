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

		// 2. Obter o utilizador, o seu cargo e as suas permissões, tudo de uma vez.
		// Graças ao nosso pré-carregamento aninhado, esta única chamada traz tudo.
		user, err := usuarioService.FindByID(userID, empresaID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Acesso negado: utilizador não encontrado."})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar permissões."})
			return
		}

		// 3. Verificar se a permissão necessária está na lista de permissões do cargo.
		hasPermission := false
		for _, p := range user.Cargo.Permissoes {
			if p.Nome == requiredPermission {
				hasPermission = true
				break // Encontrámos a permissão, não precisamos de procurar mais.
			}
		}

		// 4. Tomar a decisão final.
		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Acesso negado. O seu cargo não tem permissão para executar esta ação."})
			return
		}

		// Se o utilizador tem a permissão, deixamo-lo continuar para o handler final.
		c.Next()
	}
}
