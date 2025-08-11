package ponto

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PontoHandler struct {
	service PontoService
}

func NewPontoHandler(service PontoService) *PontoHandler {
	return &PontoHandler{
		service: service,
	}
}

func (h *PontoHandler) BaterPonto(c *gin.Context) {
	valorIDToken, existe := c.Get("userID")
	if !existe {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID do usuário não encontrado no contexto"})
		return
	}
	idTokenString, ok := valorIDToken.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID do usuário no contexto está em formato inválido"})
		return
	}
	usuarioID, err := strconv.ParseUint(idTokenString, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID do usuário no token é inválido"})
		return
	}

	valorEmpresaID, existe := c.Get("empresaID")
	if !existe {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID da empresa não encontrado no contexto"})
		return
	}
	idEmpresaString, ok := valorEmpresaID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID da empresa no contexto está em formato inválido"})
		return
	}
	empresaID, err := strconv.ParseUint(idEmpresaString, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID da empresa no token é inválido"})
		return
	}
	type BaterPontoRequest struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	var requisicao BaterPontoRequest
	if err := c.ShouldBindJSON(&requisicao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corpo da requisição (JSON) inválido"})
		return
	}

	pontoRegistrado, err := h.service.BaterPonto(uint(usuarioID), uint(empresaID), requisicao.Latitude, requisicao.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao registrar o ponto"})
		return
	}

	c.JSON(http.StatusCreated, pontoRegistrado)
}
