package auth

import (
	"net/http"

	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/gin-gonic/gin"
)

func RoleAuthMiddleware(usuarioRepo usuario.UsuarioRepository, funcoesService funcoes.FuncoesInterface, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {

		idEmpresaToken, err := funcoesService.GetUintIDFromContext(c, "empresaID")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Falha ao verificar permissões: " + err.Error()})
			return
		}

		idUsuario, err := funcoesService.GetUintIDFromContext(c, "userID")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Falha ao verificar permissões: " + err.Error()})
			return
		}

		usuario, err := usuarioRepo.FindByID(idUsuario, idEmpresaToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Acesso negado."})
			return
		}

		if usuario.Cargo.Nome != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Acesso negado. Seu cargo não tem permissão para executar esta ação."})
			return
		}
		c.Next()
	}
}
