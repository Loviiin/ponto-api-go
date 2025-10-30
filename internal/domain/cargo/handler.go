package cargo

import (
	"errors"
	"net/http"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CargoHandler struct {
	service   CargoService
	converter funcoes.FuncoesInterface
}

func NewCargoHandler(s CargoService, f funcoes.FuncoesInterface) *CargoHandler {
	return &CargoHandler{
		service:   s,
		converter: f,
	}
}

type createRequest struct {
	Nome            string `json:"nome" binding:"required,min=2,max=100" example:"Coordenador"`
	NivelHierarquia uint   `json:"nivel_hierarquia" example:"50"`
}

type updateRequest struct {
	Nome                      string  `json:"nome" binding:"omitempty,min=2,max=100" example:"Coordenador Senior"`
	NivelHierarquia           uint    `json:"nivel_hierarquia" example:"60"`
	SalarioMinimo             float64 `json:"salario_minimo" example:"3000"`
	SalarioMaximo             float64 `json:"salario_maximo" example:"5000"`
	CargaHorariaDiariaMinutos uint    `json:"carga_horaria_diaria_minutos" example:"480"`
	EntradaEsperadaMinutos    uint    `json:"entrada_esperada_minutos" example:"480"`
	SaidaEsperadaMinutos      uint    `json:"saida_esperada_minutos" example:"1020"`
	MinutosAlmocoEsperado     uint    `json:"minutos_almoco_esperado" example:"60"`
}

// @Summary      Cria um novo cargo
// @Description  Cria um novo cargo para a empresa do usuário logado. Requer permissão 'GERENCIAR_CARGOS'.
// @Tags         Cargos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        cargo  body      createRequest  true  "Dados do novo cargo"
// @Success      201    {object}  model.Cargo
// @Failure      400    {object}  map[string]string
// @Failure      409    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /cargos [post]
func (h *CargoHandler) CreateCargo(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "O corpo da requisição é inválido.",
			"details": "O campo 'nome' é obrigatório e deve ter entre 2 e 100 caracteres.",
		})
		return
	}

	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não foi possível identificar a empresa do usuário."})
		return
	}

	// Validação adicional: verificar duplicação de nome dentro da empresa
	existingCargo, _ := h.service.FindByName(req.Nome, empresaID)
	if existingCargo != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Cargo já existe.",
			"details": "Um cargo com este nome já existe nesta empresa.",
		})
		return
	}

	cargo := model.Cargo{
		Nome:            req.Nome,
		EmpresaID:       empresaID,
		NivelHierarquia: req.NivelHierarquia,
	}

	if err := h.service.Create(&cargo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Falha ao criar o cargo.",
			"details": "Ocorreu um erro ao processar a sua solicitação. Por favor, tente novamente.",
		})
		return
	}

	c.JSON(http.StatusCreated, cargo)
}

// @Summary      Lista os cargos da empresa
// @Description  Retorna uma lista de todos os cargos da empresa do usuário logado.
// @Tags         Cargos
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Cargo  "Lista de cargos"
// @Example 200 [{"id":10,"nome":"Desenvolvedor","empresa_id":1,"nivel_hierarquia":50}]
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /cargos [get]
func (h *CargoHandler) GetAllCargos(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Não foi possível identificar a empresa do usuário.",
		})
		return
	}

	cargos, err := h.service.GetAllByEmpresaID(empresaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Falha ao buscar os cargos.",
			"details": "Ocorreu um erro ao consultar a base de dados. Por favor, tente novamente.",
		})
		return
	}

	// Se nenhum cargo foi encontrado, retornar array vazio (não é erro)
	if cargos == nil {
		cargos = []model.Cargo{}
	}

	c.JSON(http.StatusOK, cargos)
}

// @Summary      Busca um cargo por ID
// @Description  Retorna os dados de um cargo específico da empresa do usuário logado.
// @Tags         Cargos
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID do Cargo"
// @Success      200  {object}  model.Cargo
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /cargos/{id} [get]
func (h *CargoHandler) GetCargoByID(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Não foi possível identificar a empresa do usuário.",
		})
		return
	}

	cargoID, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID do cargo inválido.",
			"details": "O ID deve ser um número inteiro positivo.",
		})
		return
	}

	cargo, err := h.service.FindByID(cargoID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Cargo não encontrado.",
				"details": "Nenhum cargo com este ID foi encontrado nesta empresa.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Falha ao buscar o cargo.",
			"details": "Ocorreu um erro ao consultar a base de dados. Por favor, tente novamente.",
		})
		return
	}

	c.JSON(http.StatusOK, cargo)
}

// @Summary      Atualiza um cargo
// @Description  Atualiza os dados de um cargo. Requer permissão 'GERENCIAR_CARGOS'.
// @Tags         Cargos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      int            true  "ID do Cargo"
// @Param        cargo  body      updateRequest  true  "Dados para atualização"
// @Success      200    {object}  model.Cargo
// @Failure      400    {object}  map[string]string
// @Failure      403    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      409    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /cargos/{id} [put]
func (h *CargoHandler) UpdateCargo(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Não foi possível identificar a empresa do usuário.",
		})
		return
	}

	cargoID, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID do cargo inválido.",
			"details": "O ID deve ser um número inteiro positivo.",
		})
		return
	}

	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "O corpo da requisição é inválido.",
			"details": "Verifique os tipos de dados e comprimentos dos campos (nome: 2-100 caracteres).",
		})
		return
	}

	// Verificar se há pelo menos um campo para atualizar
	if req.Nome == "" && req.NivelHierarquia == 0 && req.SalarioMinimo == 0 &&
		req.SalarioMaximo == 0 && req.CargaHorariaDiariaMinutos == 0 &&
		req.EntradaEsperadaMinutos == 0 && req.SaidaEsperadaMinutos == 0 &&
		req.MinutosAlmocoEsperado == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Nenhum campo para atualizar foi fornecido.",
			"details": "Forneça pelo menos um campo válido para atualização.",
		})
		return
	}

	// Se nome foi informado, verificar se já existe outro cargo com este nome na empresa
	if req.Nome != "" {
		existingCargo, _ := h.service.FindByName(req.Nome, empresaID)
		if existingCargo != nil && existingCargo.ID != cargoID {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Cargo com este nome já existe.",
				"details": "Um outro cargo com este nome já existe nesta empresa.",
			})
			return
		}
	}

	// Construir mapa de dados para atualização
	dados := make(map[string]interface{})
	if req.Nome != "" {
		dados["nome"] = req.Nome
	}
	if req.NivelHierarquia > 0 {
		dados["nivel_hierarquia"] = req.NivelHierarquia
	}
	if req.SalarioMinimo > 0 {
		dados["salario_minimo"] = req.SalarioMinimo
	}
	if req.SalarioMaximo > 0 {
		dados["salario_maximo"] = req.SalarioMaximo
	}
	if req.CargaHorariaDiariaMinutos > 0 {
		dados["carga_horaria_diaria_minutos"] = req.CargaHorariaDiariaMinutos
	}
	if req.EntradaEsperadaMinutos > 0 {
		dados["entrada_esperada_minutos"] = req.EntradaEsperadaMinutos
	}
	if req.SaidaEsperadaMinutos > 0 {
		dados["saida_esperada_minutos"] = req.SaidaEsperadaMinutos
	}
	if req.MinutosAlmocoEsperado > 0 {
		dados["minutos_almoco_esperado"] = req.MinutosAlmocoEsperado
	}

	// Atualizar o cargo
	err = h.service.Update(cargoID, empresaID, dados)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Cargo não encontrado.",
				"details": "Nenhum cargo com este ID foi encontrado nesta empresa.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Falha ao atualizar o cargo.",
			"details": "Ocorreu um erro ao processar a atualização. Por favor, tente novamente.",
		})
		return
	}

	// Buscar e retornar o cargo atualizado
	cargoAtualizado, err := h.service.FindByID(cargoID, empresaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Cargo atualizado mas falha ao recuperar dados.",
			"details": "A atualização foi concluída, mas houve erro ao recuperar os dados atualizados.",
		})
		return
	}

	c.JSON(http.StatusOK, cargoAtualizado)
}

// @Summary      Deleta um cargo
// @Description  Deleta um cargo da empresa. Requer permissão 'GERENCIAR_CARGOS'.
// @Tags         Cargos
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "ID do Cargo"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /cargos/{id} [delete]
func (h *CargoHandler) DeleteCargo(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cargoID, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID do cargo inválido.",
			"details": "O ID deve ser um número inteiro positivo.",
		})
		return
	}

	// Verificar se existem usuários associados ao cargo
	hasUsuarios, err := h.service.HasUsuarios(cargoID, empresaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Falha ao verificar dependências do cargo.",
			"details": "Erro ao verificar se existem funcionários associados ao cargo.",
		})
		return
	}

	if hasUsuarios {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Não é possível excluir este cargo.",
			"details": "Existem funcionários associados a este cargo. Reatribua os funcionários antes de excluir.",
		})
		return
	}

	err = h.service.Delete(cargoID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Cargo não encontrado nesta empresa.",
				"details": "O cargo especificado não existe ou não pertence a esta empresa.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Falha ao deletar o cargo.",
			"details": "Erro ao tentar deletar o cargo do banco de dados.",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      Adiciona permissão a um cargo
// @Description  Associa uma permissão existente a um cargo. Requer permissão 'GERENCIAR_CARGOS'.
// @Tags         Cargos
// @Produce      json
// @Security     BearerAuth
// @Param        id           path  int  true  "ID do Cargo"
// @Param        permissaoId  path  int  true  "ID da Permissão"
// @Success      204          "No Content"
// @Failure      400          {object}  map[string]string
// @Failure      403          {object}  map[string]string
// @Failure      404          {object}  map[string]string
// @Router       /cargos/{id}/permissoes/{permissaoId} [post]
func (h *CargoHandler) AddPermissionToCargo(c *gin.Context) {
	// Precisamos de obter a empresaID do token para garantir a segurança.
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cargoID, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do cargo inválido."})
		return
	}

	permissaoID, err := h.converter.StrParaUint(c.Param("permissaoId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da permissão inválido."})
		return
	}

	// AQUI ESTÁ A CORREÇÃO: Enviamos os 3 argumentos que o serviço espera.
	err = h.service.AddPermissionToCargo(cargoID, permissaoID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cargo ou Permissão não encontrado."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao adicionar permissão ao cargo."})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      Remove permissão de um cargo
// @Description  Remove uma permissão específica de um cargo da empresa. Requer permissão GERENCIAR_CARGOS.
// @Tags         Cargos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path  int  true  "ID do Cargo"
// @Param        permissaoId  path  int  true  "ID da Permissão"
// @Success      204          "No Content"
// @Failure      400          {object}  map[string]string
// @Failure      403          {object}  map[string]string
// @Failure      404          {object}  map[string]string
// @Router       /cargos/{id}/permissoes/{permissaoId} [delete]
func (h *CargoHandler) RemovePermissionFromCargo(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cargoID, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do cargo inválido."})
		return
	}

	permissaoID, err := h.converter.StrParaUint(c.Param("permissaoId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da permissão inválido."})
		return
	}

	err = h.service.RemovePermissionFromCargo(cargoID, permissaoID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cargo ou Permissão não encontrado."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao remover permissão do cargo."})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      Lista permissões de um cargo
// @Description  Retorna todas as permissões associadas a um cargo específico.
// @Tags         Cargos
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int  true  "ID do Cargo"
// @Success      200 {array}   model.Permissao
// @Failure      400 {object}  map[string]string
// @Failure      404 {object}  map[string]string
// @Router       /cargos/{id}/permissoes [get]
func (h *CargoHandler) GetPermissionsByCargo(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cargoID, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do cargo inválido."})
		return
	}

	permissoes, err := h.service.GetPermissionsByCargo(cargoID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cargo não encontrado."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar permissões do cargo."})
		return
	}

	c.JSON(http.StatusOK, permissoes)
}
