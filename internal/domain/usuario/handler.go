package usuario

import (
	"errors"
	"net/http"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/Loviiin/ponto-api-go/pkg/permissions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UsuarioHandler struct {
	service   UsuarioService
	converter funcoes.FuncoesInterface
}

func NewUsuarioHandler(s UsuarioService, f funcoes.FuncoesInterface) *UsuarioHandler {
	return &UsuarioHandler{
		service:   s,
		converter: f,
	}
}

type CriarUsuarioRequest struct {
	// Dados do Usuário (Pessoa)
	Nome  string `json:"nome" binding:"required" example:"Fulano de Tal"`
	CPF   string `json:"cpf" binding:"required" example:"12345678901"`
	Email string `json:"email" binding:"required,email" example:"fulano@empresa.com"`
	Senha string `json:"senha" binding:"required,min=6" example:"senha123"`

	// Dados do Contrato
	EmpresaID    uint      `json:"empresa_id" binding:"required" example:"1"`
	LocalidadeID uint      `json:"localidade_id" binding:"required" example:"10"`
	CargoID      uint      `json:"cargo_id" binding:"required" example:"5"`
	Salario      float64   `json:"salario" binding:"required" example:"3500"`
	DataAdmissao time.Time `json:"data_admissao" binding:"required" example:"2025-10-09T00:00:00Z"`
}

// UpdateUsuarioRequest define o corpo do pedido para atualizar um usuário.
type UpdateUsuarioRequest struct {
	Nome  string `json:"nome" example:"João da Silva"`
	Email string `json:"email" example:"joao.dasilva@empresa.com"`
	Senha string `json:"senha" example:"novaSenha456"`
}

// @Summary      Busca um usuário por ID
// @Description  Retorna os dados de um usuário específico da mesma empresa.
// @Tags         Usuários
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID do Usuário"
// @Success      200  {object}  model.Usuario
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /usuarios/{id} [get]
func (h *UsuarioHandler) GetByIdHandler(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O ID do usuário deve ser um número"})
		return
	}
	usuario, err := h.service.FindByID(id, empresaID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}
	c.JSON(http.StatusOK, usuario)
}

// @Summary      Lista todos os usuários da empresa
// @Description  Retorna uma lista de todos os usuários pertencentes à empresa do requisitante.
// @Tags         Usuários
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Usuario  "Exemplo"
// @Example 200 [{"id":1,"nome":"João da Silva","email":"joao@empresa.com"}]
// @Failure      500  {object}  map[string]string
// @Router       /usuarios [get]
func (h *UsuarioHandler) GetAllUsuariosHandler(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	usuarios, err := h.service.GetAll(empresaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar usuários"})
		return
	}
	c.JSON(http.StatusOK, usuarios)
}

// @Summary      Deleta um usuário
// @Description  Deleta um usuário. Requer permissão de 'DELETAR_USUARIO' ou 'DELETAR_PROPRIA_CONTA'.
// @Tags         Usuários
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID do Usuário a ser deletado"
// @Success      204  "No Content"
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /usuarios/{id} [delete]
func (h *UsuarioHandler) DeleteHandler(c *gin.Context) {
	empresaID, _ := h.converter.GetUintIDFromContext(c, "empresaID")
	idToken, _ := h.converter.GetUintIDFromContext(c, "userID")
	idUrl, _ := h.converter.StrParaUint(c.Param("id"))

	requester, err := h.service.FindByID(idToken, empresaID)
	if err != nil || requester.Contrato.ID == 0 || requester.Contrato.Cargo.ID == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado. Não foi possível verificar as suas permissões."})
		return
	}

	podeDeletar := false
	if idUrl == idToken {
		for _, p := range requester.Contrato.Cargo.Permissoes {
			if p.Nome == permissions.DELETAR_PROPRIA_CONTA {
				podeDeletar = true
				break
			}
		}
	} else {
		for _, p := range requester.Contrato.Cargo.Permissoes {
			if p.Nome == permissions.DELETAR_USUARIO {
				podeDeletar = true
				break
			}
		}
	}

	if !podeDeletar {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado. Você não tem permissão para realizar esta ação."})
		return
	}

	err = h.service.Delete(idUrl, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado para ser deletado."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao deletar o usuário."})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      Atualiza um usuário
// @Description  Atualiza os dados de um usuário. O corpo do pedido pode conter qualquer um dos campos definidos no modelo.
// @Tags         Usuários
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                  true  "ID do Usuário a ser atualizado"
// @Param        dados    body      UpdateUsuarioRequest true  "Dados para atualização"
// @Success      204      "No Content"
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Router       /usuarios/{id} [put]
func (h *UsuarioHandler) UpdateUsuarioHandler(c *gin.Context) {
	empresaID, _ := h.converter.GetUintIDFromContext(c, "empresaID")
	idToken, _ := h.converter.GetUintIDFromContext(c, "userID")
	idUrl, _ := h.converter.StrParaUint(c.Param("id"))

	requester, err := h.service.FindByID(idToken, empresaID)
	if err != nil || requester.Contrato.ID == 0 || requester.Contrato.Cargo.ID == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado. Não foi possível verificar as suas permissões."})
		return
	}

	podeEditar := false
	if idUrl == idToken {
		for _, p := range requester.Contrato.Cargo.Permissoes {
			if p.Nome == permissions.EDITAR_PROPRIA_CONTA {
				podeEditar = true
				break
			}
		}
	} else {
		for _, p := range requester.Contrato.Cargo.Permissoes {
			if p.Nome == permissions.EDITAR_USUARIO {
				podeEditar = true
				break
			}
		}
	}

	if !podeEditar {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para editar este usuário."})
		return
	}

	var dadosParaAtualizar map[string]interface{}
	if err := c.ShouldBindJSON(&dadosParaAtualizar); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corpo da requisição (JSON) inválido"})
		return
	}

	// Remove campos que não devem ser atualizados diretamente nesta rota
	delete(dadosParaAtualizar, "cargo_id")
	delete(dadosParaAtualizar, "empresa_id")
	delete(dadosParaAtualizar, "localidade_id")
	delete(dadosParaAtualizar, "salario")
	delete(dadosParaAtualizar, "data_admissao")

	err = h.service.Update(idUrl, empresaID, dadosParaAtualizar)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar o usuário."})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      Cria um novo usuário
// @Description  Cria um novo usuário (funcionário) e seu contrato de trabalho no sistema.
// @Description  Regras: o requisitante só pode atribuir cargos com nível hierárquico menor ou igual ao seu.
// @Tags         Usuários
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        usuario  body      CriarUsuarioRequest  true  "Dados do Novo Usuário e Contrato"
// @Example      {"nome":"Fulano de Tal","cpf":"12345678901","email":"fulano@empresa.com","senha":"senha123","empresa_id":1,"localidade_id":10,"cargo_id":5,"salario":3500,"data_admissao":"2025-10-09T00:00:00Z"}
// @Success      201      {object}  model.Usuario
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Router       /usuarios [post]
func (h *UsuarioHandler) CriarUsuarioHandler(c *gin.Context) {
	var request CriarUsuarioRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Obter ID do requisitante do contexto
	idRequisitante, err := h.converter.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Falha ao identificar o requisitante."})
		return
	}

	// Força o empresaID a ser o do token para evitar uso indevido ou inconsistências no payload
	empresaIDToken, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Falha ao identificar a empresa do requisitante."})
		return
	}

	usuario := &model.Usuario{
		Nome:  request.Nome,
		CPF:   request.CPF,
		Email: request.Email,
		Senha: request.Senha,
	}

	contrato := &model.Contrato{
		// Ignora o empresa_id do payload e utiliza o do token
		EmpresaID:    empresaIDToken,
		LocalidadeID: request.LocalidadeID,
		CargoID:      request.CargoID,
		Salario:      request.Salario,
		DataAdmissao: request.DataAdmissao,
	}

	err = h.service.CriarUsuarioEContrato(usuario, contrato, idRequisitante)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, usuario)
}

// @Summary      Obtém os dados do usuário logado
// @Description  Retorna as informações detalhadas do usuário que está a fazer o pedido.
// @Tags         Usuários
// @Produce      json
// @Success      200  {object}  model.Usuario
// @Example 200 {"id":1,"nome":"Dono","email":"dono@ponto.com","contrato":{"cargo":{"id":1,"nome":"Dono","nivel_hierarquia":100}}}
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /usuarios/me [get]
func (h *UsuarioHandler) GetMeuPerfil(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := h.converter.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	usuario, err := h.service.FindByID(id, empresaID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	c.JSON(http.StatusOK, usuario)
}
