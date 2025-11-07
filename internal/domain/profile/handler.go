package profile

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Loviiin/ponto-api-go/pkg/cloudinary"
	"github.com/gin-gonic/gin"
)

// Handler define a interface para os handlers de perfil
type Handler struct {
	service           Service
	cloudinaryService cloudinary.Service
}

// NewHandler cria uma nova instância do handler
func NewHandler(service Service, cloudinaryService cloudinary.Service) *Handler {
	return &Handler{
		service:           service,
		cloudinaryService: cloudinaryService,
	}
}

// Helper function to convert string userID from context to uint
func getUserIDFromContext(c *gin.Context) (uint, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, fmt.Errorf("usuário não autenticado")
	}

	// userID is stored as string in the context by auth middleware
	userIDStr, ok := userID.(string)
	if !ok {
		return 0, fmt.Errorf("formato de userID inválido")
	}

	userIDUint, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("erro ao converter userID: %v", err)
	}

	return uint(userIDUint), nil
}

// Helper function to convert string empresaID from context to uint
func getEmpresaIDFromContext(c *gin.Context) (uint, error) {
	empresaID, exists := c.Get("empresaID")
	if !exists {
		return 0, fmt.Errorf("empresa não identificada")
	}

	// empresaID is stored as string in the context by auth middleware
	empresaIDStr, ok := empresaID.(string)
	if !ok {
		return 0, fmt.Errorf("formato de empresaID inválido")
	}

	empresaIDUint, err := strconv.ParseUint(empresaIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("erro ao converter empresaID: %v", err)
	}

	return uint(empresaIDUint), nil
}

// GetMyProfile godoc
// @Summary Obter meu perfil
// @Description Retorna todas as informações do perfil do usuário autenticado
// @Tags Profile
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ProfileResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profile/me [get]
func (h *Handler) GetMyProfile(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.service.GetMyProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile godoc
// @Summary Atualizar meu perfil
// @Description Atualiza informações editáveis do perfil (nome, telefone, email). Usa PATCH semântico - apenas campos enviados são atualizados. SENHA NUNCA É ALTERADA AQUI.
// @Tags Profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Dados a atualizar (apenas campos que deseja modificar)"
// @Success 200 {object} ProfileResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profile/me [patch]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos", "details": err.Error()})
		return
	}

	// Coletar IP e User-Agent para audit log
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	profile, err := h.service.UpdateProfile(userID, req, ip, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// ChangePassword godoc
// @Summary Alterar senha
// @Description Altera a senha do usuário autenticado. Este é o ÚNICO endpoint que pode modificar a senha.
// @Tags Profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Dados para alteração de senha"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profile/me/password [patch]
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos", "details": err.Error()})
		return
	}

	// Coletar IP e User-Agent para audit log
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := h.service.ChangePassword(userID, req, ip, userAgent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "senha alterada com sucesso"})
}

// GetMyStats godoc
// @Summary Obter minhas estatísticas
// @Description Retorna estatísticas de ponto, banco de horas e atividades
// @Tags Profile
// @Security BearerAuth
// @Produce json
// @Success 200 {object} StatsResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profile/me/stats [get]
func (h *Handler) GetMyStats(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	empresaID, err := getEmpresaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	stats, err := h.service.GetMyStats(userID, empresaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetMyPermissions godoc
// @Summary Obter minhas permissões
// @Description Retorna todas as permissões do usuário autenticado
// @Tags Profile
// @Security BearerAuth
// @Produce json
// @Success 200 {object} PermissoesResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profile/me/permissions [get]
func (h *Handler) GetMyPermissions(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	permissions, err := h.service.GetMyPermissions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// GetRecentActivity godoc
// @Summary Obter atividades recentes
// @Description Retorna as últimas atividades do usuário no sistema
// @Tags Profile
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limite de registros" default(20)
// @Success 200 {object} RecentActivityResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profile/me/recent-activity [get]
func (h *Handler) GetRecentActivity(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	limit := 20
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	activity, err := h.service.GetRecentActivity(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, activity)
}

// GetCalendar godoc
// @Summary Obter calendário de presença
// @Description Retorna o calendário de presença do usuário para um mês específico
// @Tags Profile
// @Security BearerAuth
// @Produce json
// @Param month query int false "Mês (1-12)" default(0)
// @Param year query int false "Ano" default(0)
// @Success 200 {object} CalendarioResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profile/me/calendar [get]
func (h *Handler) GetCalendar(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Default: mês e ano atuais
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	// Parse dos parâmetros com validação melhorada
	if monthParam := c.Query("month"); monthParam != "" {
		parsedMonth, err := strconv.Atoi(monthParam)
		if err != nil || parsedMonth < 1 || parsedMonth > 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mês deve estar entre 1 e 12", "received": monthParam})
			return
		}
		month = parsedMonth
	}

	if yearParam := c.Query("year"); yearParam != "" {
		parsedYear, err := strconv.Atoi(yearParam)
		if err != nil || parsedYear < 2000 || parsedYear > 2100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ano deve estar entre 2000 e 2100", "received": yearParam})
			return
		}
		year = parsedYear
	}

	// Debug: Log dos valores
	fmt.Printf("[DEBUG] GetCalendar - userID: %d, month: %d, year: %d\n", userID, month, year)

	calendar, err := h.service.GetCalendar(userID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, calendar)
}

// UploadAvatar godoc
// @Summary Upload de foto de perfil
// @Description Faz upload da foto de perfil do usuário para Cloudinary (CDN global)
// @Tags Profile
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Arquivo de imagem (JPEG/PNG, máx 5MB)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profile/me/avatar [post]
func (h *Handler) UploadAvatar(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Verificar se Cloudinary está configurado
	if h.cloudinaryService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "serviço de upload não está configurado. Configure CLOUDINARY_URL no .env"})
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "arquivo não enviado"})
		return
	}
	defer file.Close()

	// Validar tipo de arquivo
	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/jpg" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "apenas imagens JPEG e PNG são permitidas"})
		return
	}

	// Validar tamanho (máximo 5MB)
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "arquivo muito grande. Máximo: 5MB"})
		return
	}

	// Obter avatar antigo para deletar depois
	profile, err := h.service.GetMyProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao obter perfil atual"})
		return
	}

	// Upload para Cloudinary
	avatarURL, err := h.cloudinaryService.UploadAvatar(file, userID, header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("erro ao fazer upload: %v", err)})
		return
	}

	// Atualizar campo Avatar no banco de dados
	if err := h.service.UpdateAvatar(userID, avatarURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao atualizar avatar no perfil"})
		return
	}

	// Deletar avatar antigo do Cloudinary (se existir)
	if profile.Avatar != "" {
		oldPublicID := cloudinary.ExtractPublicIDFromURL(profile.Avatar)
		if oldPublicID != "" {
			// Deletar de forma assíncrona (não bloqueia a resposta)
			go func() {
				_ = h.cloudinaryService.DeleteAvatar(oldPublicID)
			}()
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "avatar atualizado com sucesso",
		"url":     avatarURL,
	})
}

// UpdateCPF godoc
// @Summary Atualizar CPF de um usuário (Admin)
// @Description Permite que um admin com permissão EDITAR_USUARIO atualize o CPF de qualquer usuário
// @Tags Profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path int true "ID do usuário"
// @Param request body UpdateCPFRequest true "Novo CPF"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/users/{user_id}/cpf [patch]
func (h *Handler) UpdateCPF(c *gin.Context) {
	adminID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Obter ID do usuário alvo
	targetUserIDStr := c.Param("user_id")
	targetUserIDUint, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}
	targetUserID := uint(targetUserIDUint)

	var req UpdateCPFRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos", "details": err.Error()})
		return
	}

	// Coletar IP e User-Agent para audit log
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := h.service.UpdateCPF(adminID, targetUserID, req, ip, userAgent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "CPF atualizado com sucesso"})
}
