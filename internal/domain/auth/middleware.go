package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Loviiin/ponto-api-go/pkg/jwt" // Importa o nosso serviço de JWT
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtService *jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token de autorização não fornecido"})
			return
		}

		splitToken := strings.Split(authHeader, " ")
		if len(splitToken) != 2 || splitToken[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Formato do token de autorização inválido"})
			return
		}

		tokenString := splitToken[1]
		token, err := jwtService.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token de autorização inválido ou expirado"})
			return
		}

		claims, ok := token.Claims.(*jwt.PontoClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Não foi possível processar as claims do token"})
			return
		}

		userID := claims.Subject

		empresaID := fmt.Sprintf("%d", claims.EmpresaID)

		c.Set("userID", userID)
		c.Set("empresaID", empresaID)

		c.Next()
	}
}
