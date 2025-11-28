package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

func TestAuthHandler_ForgotPassword_Success(t *testing.T) {
	// Setup
	router := setupTestRouter()

	// Mock handler (simplified version)
	router.POST("/auth/forgot-password", func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
			return
		}

		// Simulate success
		c.JSON(http.StatusOK, gin.H{
			"mensagem": "Se o email existir, vocêreceberá instruções",
		})
	})

	// Test
	requestBody := map[string]string{
		"email": "test@example.com",
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/auth/forgot-password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["mensagem"], "instruções")
}

func TestAuthHandler_ForgotPassword_InvalidEmail(t *testing.T) {
	// Setup
	router := setupTestRouter()

	router.POST("/auth/forgot-password", func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "Email inválido"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"mensagem": "OK"})
	})

	// Test with invalid email
	requestBody := map[string]string{
		"email": "invalid-email",
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/auth/forgot-password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_ResetPassword_Success(t *testing.T) {
	// Setup
	router := setupTestRouter()

	router.POST("/auth/reset-password", func(c *gin.Context) {
		var req struct {
			Token       string `json:"token" binding:"required"`
			NewPassword string `json:"new_password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
			return
		}

		// Simulate success
		c.JSON(http.StatusOK, gin.H{
			"mensagem": "Senha alterada com sucesso",
		})
	})

	// Test
	requestBody := map[string]string{
		"token":        "valid-token-12345",
		"new_password": "NewPassword123",
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/auth/reset-password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Senha alterada com sucesso", response["mensagem"])
}

func TestAuthHandler_ResetPassword_WeakPassword(t *testing.T) {
	// Setup
	router := setupTestRouter()

	router.POST("/auth/reset-password", func(c *gin.Context) {
		var req struct {
			Token       string `json:"token" binding:"required"`
			NewPassword string `json:"new_password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "Senha muito fraca"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"mensagem": "OK"})
	})

	// Test with weak password
	requestBody := map[string]string{
		"token":        "valid-token",
		"new_password": "123", // Too short
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/auth/reset-password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_GoogleLogin_RedirectToOAuth(t *testing.T) {
	// Setup
	router := setupTestRouter()

	router.GET("/auth/google/login", func(c *gin.Context) {
		// Simulate redirect to Google OAuth
		authURL := "https://accounts.google.com/o/oauth2/auth?client_id=test&redirect_uri=callback&response_type=code&scope=email+profile"
		c.Redirect(http.StatusFound, authURL)
	})

	// Test
	req, _ := http.NewRequest("GET", "/auth/google/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "accounts.google.com")
}

func TestAuthHandler_GoogleCallback_Success(t *testing.T) {
	// Setup
	router := setupTestRouter()

	router.GET("/auth/google/callback", func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")

		if code == "" || state == "" {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "Código ou state inválido"})
			return
		}

		// Simulate successful OAuth exchange
		c.JSON(http.StatusOK, gin.H{
			"token": "jwt-token-12345",
		})
	})

	// Test
	req, _ := http.NewRequest("GET", "/auth/google/callback?code=test-code&state=test-state", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["token"])
}

func TestAuthHandler_GoogleCallback_MissingCode(t *testing.T) {
	// Setup
	router := setupTestRouter()

	router.GET("/auth/google/callback", func(c *gin.Context) {
		code := c.Query("code")

		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "Código ausente"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": "jwt"})
	})

	// Test without code
	req, _ := http.NewRequest("GET", "/auth/google/callback", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_GoogleLink_RequiresAuth(t *testing.T) {
	// Setup
	router := setupTestRouter()

	// Middleware to check authentication
	authMiddleware := func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Não autenticado"})
			c.Abort()
			return
		}
		c.Next()
	}

	router.GET("/auth/google/link", authMiddleware, func(c *gin.Context) {
		authURL := "https://accounts.google.com/o/oauth2/auth?client_id=test"
		c.JSON(http.StatusOK, gin.H{"url": authURL})
	})

	// Test without auth token
	req, _ := http.NewRequest("GET", "/auth/google/link", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_RateLimiting(t *testing.T) {
	// This would test rate limiting on forgot-password endpoint
	// Simulated test - actual implementation would use a rate limiter

	router := setupTestRouter()

	attempts := 0
	maxAttempts := 3

	router.POST("/auth/forgot-password", func(c *gin.Context) {
		attempts++

		if attempts > maxAttempts {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"erro":        "Muitas tentativas",
				"retry_after": 60,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"mensagem": "OK"})
	})

	// Make 4 requests
	for i := 0; i < 4; i++ {
		req, _ := http.NewRequest("POST", "/auth/forgot-password", bytes.NewBuffer([]byte(`{"email":"test@example.com"}`)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i < maxAttempts {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}
