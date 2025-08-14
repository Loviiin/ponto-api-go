package permissao

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Create(c *gin.Context) {
	var permissao model.Permissao
	if err := c.ShouldBindJSON(&permissao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "corpo da requisição inválido"})
		return
	}

	if err := h.service.Create(&permissao); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao criar permissão"})
		return
	}
	c.JSON(http.StatusCreated, permissao)
}

func (h *Handler) FindAll(c *gin.Context) {
	permissoes, err := h.service.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar permissões"})
		return
	}
	c.JSON(http.StatusOK, permissoes)
}
