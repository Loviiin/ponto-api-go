package empresa

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Loviiin/ponto-api-go/internal/config"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// A struct agora é mais simples, sem a dependência do serviço de usuário.
type EmpresaHandler struct {
	service   EmpresaService
	converter funcoes.FuncoesInterface
	Db        *gorm.DB // NOVA LINHA
}

func NewEmpresaHandler(s EmpresaService, f funcoes.FuncoesInterface, db *gorm.DB) *EmpresaHandler { // NOVO PARÂMETRO
	return &EmpresaHandler{
		service:   s,
		converter: f,
		Db:        db, // NOVA LINHA
	}
}

type criaEmpresaRequest struct {
	NomeFantasia string `json:"nomeFantasia" binding:"required" example:"Minha Empresa"`
	RazaoSocial  string `json:"razaoSocial"  binding:"required" example:"Minha Empresa LTDA"`
	CNPJ         string `json:"cnpj"         binding:"required" example:"12.345.678/0001-95"`
}

// @Summary      Cria uma nova empresa
// @Description  Registra uma nova empresa no sistema e configura cargos e permissões padrão para ela.
// @Tags         Empresas
// @Accept       json
// @Produce      json
// @Param        empresa  body      criaEmpresaRequest  true  "Dados da nova empresa"
// @Success      201      {object}  model.Empresa
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /empresas [post]
func (h *EmpresaHandler) CriarEmpresaHandler(c *gin.Context) {

	var request criaEmpresaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Normalização básica dos campos e validação simples de CNPJ
	request.NomeFantasia = strings.TrimSpace(request.NomeFantasia)
	request.RazaoSocial = strings.TrimSpace(request.RazaoSocial)
	// Normalizar CNPJ: manter apenas dígitos
	cnpjDigits := make([]rune, 0, len(request.CNPJ))
	for _, r := range request.CNPJ {
		if r >= '0' && r <= '9' {
			cnpjDigits = append(cnpjDigits, r)
		}
	}
	normalizedCNPJ := string(cnpjDigits)
	if request.NomeFantasia == "" || request.RazaoSocial == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nome fantasia e razão social são obrigatórios."})
		return
	}
	if len(normalizedCNPJ) != 14 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CNPJ inválido. Informe 14 dígitos."})
		return
	}

	empresa := model.Empresa{
		NomeFantasia: request.NomeFantasia,
		RazaoSocial:  request.RazaoSocial,
		CNPJ:         normalizedCNPJ,
	}

	err := h.service.CreateEmpresa(&empresa)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	permissoes := config.SeedPermissions(h.Db)
	config.SetupDefaultRolesAndPermissions(h.Db, empresa.ID, permissoes)
	c.JSON(http.StatusCreated, empresa)
}

// @Summary      Lista todas as empresas
// @Description  Retorna uma lista de todas as empresas cadastradas. Requer autenticação.
// @Tags         Empresas
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Empresa
// @Failure      500  {object}  map[string]string
// @Router       /empresas [get]
func (h *EmpresaHandler) GetAllEmpresasHandler(c *gin.Context) {
	empresas, err := h.service.GetAllEmpresasSer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar empresas"})
		return
	}
	c.JSON(http.StatusOK, empresas)
}

// @Summary      Busca uma empresa por ID
// @Description  Retorna os dados de uma empresa específica pelo seu ID.
// @Tags         Empresas
// @Produce      json
// @Param        id   path      int  true  "ID da Empresa"
// @Success      200  {object}  model.Empresa
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /empresas/{id} [get]
func (h *EmpresaHandler) GetEmpresaByIDHandler(c *gin.Context) {
	id, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID da empresa deve ser um número válido"})
		return
	}

	empresa, err := h.service.GetEmpresaByIDSer(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Empresa não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar empresa"})
		return
	}
	c.JSON(http.StatusOK, empresa)
}

// @Summary      Atualiza uma empresa
// @Description  Atualiza os dados de uma empresa. Requer permissão 'EDITAR_EMPRESA'.
// @Tags         Empresas
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      int                  true  "ID da Empresa"
// @Param        dados  body      map[string]interface{}  true  "Dados para atualização"
// @Success      204    "No Content"
// @Failure      400    {object}  map[string]string
// @Failure      403    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /empresas/{id} [put]
func (h *EmpresaHandler) UpdateEmpresaHandler(c *gin.Context) {
	idEmpresa, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID da empresa na URL é inválido."})
		return
	}

	var dadosParaAtualizar map[string]interface{}
	if err := c.ShouldBindJSON(&dadosParaAtualizar); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corpo da requisição (JSON) inválido"})
		return
	}

	if err := h.service.UpdateEmpresaSer(idEmpresa, dadosParaAtualizar); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar a empresa"})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      Deleta uma empresa
// @Description  Deleta uma empresa permanentemente. Requer permissão 'DELETAR_EMPRESA'.
// @Tags         Empresas
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "ID da Empresa"
// @Success      204 "No Content"
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /empresas/{id} [delete]
func (h *EmpresaHandler) DeleteEmpresaHandler(c *gin.Context) {
	idEmpresa, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID da empresa na URL é inválido."})
		return
	}

	err = h.service.DeleteEmpresaSer(idEmpresa)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Empresa não encontrada para deletar."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao deletar a empresa."})
		return
	}

	c.Status(http.StatusNoContent)
}
