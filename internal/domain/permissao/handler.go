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

// @Summary      Cria uma nova permissão
// @Description  (Admin) Cria uma nova permissão global no sistema.
// @Tags         Permissões
// @Accept       json
// @Produce      json
// @Param        permissao  body      model.Permissao  true  "Dados da nova permissão"
// @Success      201        {object}  model.Permissao
// @Failure      400        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /permissoes [post]
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

// @Summary      Lista todas as permissões
// @Description  (Admin) Retorna uma lista de todas as permissões disponíveis no sistema.
// @Tags         Permissões
// @Produce      json
// @Success      200  {array}   model.Permissao
// @Failure      500  {object}  map[string]string
// @Router       /permissoes [get]
func (h *Handler) FindAll(c *gin.Context) {
	permissoes, err := h.service.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar permissões"})
		return
	}
	c.JSON(http.StatusOK, permissoes)
}
