package cep

import (
	"errors"
	"net/http"

	"github.com/Loviiin/ponto-api-go/pkg/cep"
	"github.com/Loviiin/ponto-api-go/pkg/geolocation"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service geolocation.Service
}

func NewHandler(s geolocation.Service) *Handler { return &Handler{service: s} }

// GetByCEPRequest path parameter is `cep`

// @Summary       Consulta CEP
// @Description   Retorna o endereço unificado a partir do CEP informado (8 dígitos, com ou sem hífen), enriquecendo com latitude/longitude quando necessário.
// @Tags          CEP
// @Produce       json
// @Param         cep  path      string  true  "CEP"  example:"01001-000"
// @Success       200  {object}  model.Localidade
// @Failure       400  {object}  map[string]string
// @Failure       500  {object}  map[string]string
// @Router        /cep/v2/{cep} [get]
func (h *Handler) GetByCEP(c *gin.Context) {
	cepParam := c.Param("cep")
	localidade, err := h.service.GetLocationFromCEP(cepParam)
	if err != nil {
		if errors.Is(err, cep.ErrInvalidCEP) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao consultar CEP: " + err.Error()})
		return
	}
	// Sucesso
	c.JSON(http.StatusOK, localidade)
}

// @Summary       Busca por CEP com múltiplos providers e geolocalização.
// @Description   Versão 2 do serviço de busca por CEP que inclui latitude e longitude além dos dados básicos. Utiliza múltiplos provedores e enriquecimento via geolocalização.
// @Tags          CEP V2
// @Produce       json
// @Param         cep  path      string  true  "CEP"  example:"01310-930"
// @Success       200  {object}  model.Localidade
// @Failure       400  {object}  map[string]string
// @Failure       500  {object}  map[string]string
// @Router        /cep/v2/{cep} [get]
func (h *Handler) GetByCEPV2(c *gin.Context) {
	h.GetByCEP(c)
}
