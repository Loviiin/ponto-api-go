package localidade

import (
	"errors"
	"net/http"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/cep"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   Service
	converter funcoes.FuncoesInterface
}

func NewHandler(s Service, f funcoes.FuncoesInterface) *Handler {
	return &Handler{service: s, converter: f}
}

// Struct para o corpo da requisição de criação de localidade
type createRequest struct {
	Nome               string  `json:"nome" binding:"required"`
	CEP                string  `json:"cep" binding:"required"`
	EmpresaID          uint    `json:"empresa_id" binding:"required"`
	RaioGeofenceMetros float64 `json:"raio_geofence_metros" binding:"required"`
}

// @Summary      Cria uma nova localidade
// @Description  Cria uma nova localidade (matriz ou filial) para uma empresa, buscando o endereço e as coordenadas a partir do CEP (CEP deve ter 8 dígitos). Requer permissão 'GERENCIAR_LOCALIDADES'.
// @Tags         Localidades
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        localidade body      createRequest       true  "Dados da nova localidade"
// @Success      201        {object}  model.Localidade
// @Failure      400        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /localidades [post]
func (h *Handler) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corpo da requisição inválido: " + err.Error()})
		return
	}

	localidadeParcial := &model.Localidade{
		Nome:               req.Nome,
		CEP:                req.CEP,
		EmpresaID:          req.EmpresaID,
		RaioGeofenceMetros: req.RaioGeofenceMetros,
	}

	if err := h.service.Create(localidadeParcial); err != nil {
		if errors.Is(err, cep.ErrInvalidCEP) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar localidade: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, localidadeParcial)
}

// @Summary      Lista as localidades de uma empresa
// @Description  Retorna uma lista de todas as localidades de uma empresa. Requer permissão 'GERENCIAR_LOCALIDADES'.
// @Tags         Localidades
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID da Empresa"
// @Success      200  {array}   model.Localidade
// @Failure      500  {object}  map[string]string
// @Router       /empresas/{id}/localidades [get]
func (h *Handler) GetAllByEmpresa(c *gin.Context) {
	empresaID, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de empresa inválido."})
		return
	}

	localidades, err := h.service.FindAllByEmpresaID(empresaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar localidades."})
		return
	}

	c.JSON(http.StatusOK, localidades)
}
