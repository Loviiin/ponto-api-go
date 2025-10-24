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
	service           UsuarioService
	bancoHorasService BancoHorasService // Opcional: para incluir saldo
	converter         funcoes.FuncoesInterface
}

// Interface local para incluir saldo de banco de horas
type BancoHorasService interface {
	GetSaldoAtualUsuario(usuarioID uint, empresaID uint) (int, error)
}

func NewUsuarioHandler(s UsuarioService, f funcoes.FuncoesInterface) *UsuarioHandler {
	return &UsuarioHandler{
		service:           s,
		bancoHorasService: nil, // Mantém compatibilidade
		converter:         f,
	}
}

// NewUsuarioHandlerWithBancoHoras cria handler com suporte a saldo de banco de horas
func NewUsuarioHandlerWithBancoHoras(s UsuarioService, bhs BancoHorasService, f funcoes.FuncoesInterface) *UsuarioHandler {
	return &UsuarioHandler{
		service:           s,
		bancoHorasService: bhs,
		converter:         f,
	}
}

// UsuarioComSaldo estende o modelo Usuario para incluir saldo de banco de horas
type UsuarioComSaldo struct {
	model.Usuario
	SaldoBancoHorasMinutos *int `json:"saldo_banco_horas_minutos,omitempty"`
}

type CriarUsuarioRequest struct {
	// Dados do Usuário (Pessoa)
	Nome  string `json:"nome" binding:"required,min=2,max=100" example:"Fulano de Tal"`
	CPF   string `json:"cpf" binding:"required,len=11" example:"12345678901"`
	Email string `json:"email" binding:"required,email,max=100" example:"fulano@empresa.com"`
	Senha string `json:"senha" binding:"required,min=8" example:"senha123"`

	// Dados do Contrato
	LocalidadeID uint      `json:"localidade_id" binding:"required" example:"10"`
	CargoID      uint      `json:"cargo_id" binding:"required" example:"5"`
	Salario      float64   `json:"salario" binding:"required,gt=0" example:"3500"`
	DataAdmissao time.Time `json:"data_admissao" binding:"required" example:"2025-10-09T00:00:00Z"`
}

// UpdateUsuarioRequest define o corpo do pedido para atualizar um usuário.
type UpdateUsuarioRequest struct {
	Nome  string `json:"nome" binding:"omitempty,min=2,max=100" example:"João da Silva"`
	Email string `json:"email" binding:"omitempty,email,max=100" example:"joao.dasilva@empresa.com"`
}

// PatchUsuarioRequest define o corpo do pedido para atualização parcial de um usuário (incluindo cargo).
type PatchUsuarioRequest struct {
	Nome    *string `json:"nome" binding:"omitempty,min=2,max=100" example:"João da Silva"`
	Email   *string `json:"email" binding:"omitempty,email,max=100" example:"joao.dasilva@empresa.com"`
	CargoID *uint   `json:"cargoId" binding:"omitempty" example:"5"`
}

// @Summary      Busca um usuário por ID
// @Description  Retorna os dados de um usuário específico da mesma empresa, incluindo contrato, cargo, localidade.
// @Tags         Usuários
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID do Usuário"
// @Success      200  {object}  model.Usuario
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /usuarios/{id} [get]
func (h *UsuarioHandler) GetByIdHandler(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Erro de autenticação.",
			"details": "Não foi possível identificar a empresa.",
		})
		return
	}

	id, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID do usuário inválido.",
			"details": "O ID deve ser um número inteiro positivo.",
		})
		return
	}

	usuario, err := h.service.FindByID(id, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Usuário não encontrado.",
				"details": "O usuário especificado não existe ou não pertence a esta empresa.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Falha ao buscar usuário.",
			"details": "Ocorreu um erro ao consultar o usuário no banco de dados.",
		})
		return
	}

	c.JSON(http.StatusOK, usuario)
}

// @Summary      Lista todos os usuários da empresa
// @Description  Retorna uma lista de todos os usuários pertencentes à empresa do requisitante. Suporta paginação e inclusão opcional de saldo de banco de horas.
// @Tags         Usuários
// @Produce      json
// @Security     BearerAuth
// @Param        page           query     int     false  "Número da página (padrão: 1)"
// @Param        limit          query     int     false  "Itens por página (padrão: 50, máximo: 100)"
// @Param        include_saldo  query     bool    false  "Incluir saldo de banco de horas (padrão: false)"
// @Success      200  {object}  map[string]interface{}  "Lista de usuários com metadados de paginação"
// @Example 200 {"usuarios":[{"id":1,"nome":"João da Silva","email":"joao@empresa.com","saldo_banco_horas_minutos":120,"contrato":{"cargo":{"nome":"Desenvolvedor"}}}],"page":1,"limit":50,"total":120}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /usuarios [get]
func (h *UsuarioHandler) GetAllUsuariosHandler(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erro de autenticação.",
			"details": "Não foi possível identificar a empresa.",
		})
		return
	}

	// Parâmetros de paginação
	page := 1
	limit := 50

	// Parse query params
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := h.converter.StrParaUint(pageStr); err == nil && p > 0 {
			page = int(p)
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := h.converter.StrParaUint(limitStr); err == nil && l > 0 {
			limit = int(l)
			// Limitar máximo de registros por página
			if limit > 100 {
				limit = 100
			}
		}
	}

	// Verificar se deve incluir saldo de banco de horas
	includeSaldo := c.Query("include_saldo") == "true"

	// Buscar com paginação
	usuarios, total, err := h.service.GetAllPaginated(empresaID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Falha ao buscar usuários.",
			"details": "Ocorreu um erro ao consultar os usuários no banco de dados.",
		})
		return
	}

	// Calcular metadados de paginação
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	// Se include_saldo=true e temos o service, enriquecer com saldo
	var response interface{}
	if includeSaldo && h.bancoHorasService != nil {
		usuariosComSaldo := make([]UsuarioComSaldo, 0, len(usuarios))
		for _, u := range usuarios {
			usuarioComSaldo := UsuarioComSaldo{Usuario: u}
			// Buscar saldo (com cache)
			if saldo, err := h.bancoHorasService.GetSaldoAtualUsuario(u.ID, empresaID); err == nil {
				usuarioComSaldo.SaldoBancoHorasMinutos = &saldo
			}
			usuariosComSaldo = append(usuariosComSaldo, usuarioComSaldo)
		}
		response = usuariosComSaldo
	} else {
		response = usuarios
	}

	// Retornar lista vazia se não houver usuários
	if len(usuarios) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"usuarios":    []model.Usuario{},
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": 0,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"usuarios":    response,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	})
}

// @Summary      Deleta um usuário
// @Description  Deleta um usuário (soft delete). Requer permissão de 'DELETAR_USUARIO' ou 'DELETAR_PROPRIA_CONTA'.
// @Tags         Usuários
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID do Usuário a ser deletado"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /usuarios/{id} [delete]
func (h *UsuarioHandler) DeleteHandler(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Erro de autenticação.",
			"details": "Não foi possível identificar a empresa.",
		})
		return
	}

	idToken, err := h.converter.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Erro de autenticação.",
			"details": "Não foi possível identificar o usuário requisitante.",
		})
		return
	}

	idUrl, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID do usuário inválido.",
			"details": "O ID deve ser um número inteiro positivo.",
		})
		return
	}

	// Verificar permissões
	requester, err := h.service.FindByID(idToken, empresaID)
	if err != nil || requester.Contrato.ID == 0 || requester.Contrato.Cargo.ID == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Acesso negado.",
			"details": "Não foi possível verificar as suas permissões.",
		})
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
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Acesso negado.",
			"details": "Você não tem permissão para deletar este usuário. Permissões necessárias: DELETAR_USUARIO ou DELETAR_PROPRIA_CONTA.",
		})
		return
	}

	// Deletar usuário
	err = h.service.Delete(idUrl, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Usuário não encontrado.",
				"details": "O usuário especificado não existe ou não pertence a esta empresa.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Falha ao deletar o usuário.",
			"details": "Ocorreu um erro ao processar a exclusão do usuário.",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      Atualiza um usuário
// @Description  Atualiza os dados básicos de um usuário (nome e email). Não permite atualizar senha (use endpoint específico).
// @Tags         Usuários
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                  true  "ID do Usuário a ser atualizado"
// @Param        dados    body      UpdateUsuarioRequest true  "Dados para atualização"
// @Success      200      {object}  model.Usuario        "Usuário atualizado com sucesso"
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Router       /usuarios/{id} [put]
func (h *UsuarioHandler) UpdateUsuarioHandler(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Erro de autenticação.",
			"details": "Não foi possível identificar a empresa.",
		})
		return
	}

	idToken, err := h.converter.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Erro de autenticação.",
			"details": "Não foi possível identificar o usuário requisitante.",
		})
		return
	}

	idUrl, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID do usuário inválido.",
			"details": "O ID deve ser um número inteiro positivo.",
		})
		return
	}

	// Verificar permissões
	requester, err := h.service.FindByID(idToken, empresaID)
	if err != nil || requester.Contrato.ID == 0 || requester.Contrato.Cargo.ID == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Acesso negado.",
			"details": "Não foi possível verificar as suas permissões.",
		})
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
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Você não tem permissão para editar este usuário.",
			"details": "Permissões necessárias: EDITAR_USUARIO ou EDITAR_PROPRIA_CONTA.",
		})
		return
	}

	// Parse e validar request
	var request UpdateUsuarioRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dados de requisição inválidos.",
			"details": "Verifique os campos: nome (2-100 caracteres), email (formato válido).",
		})
		return
	}

	// Verificar se há pelo menos um campo para atualizar
	if request.Nome == "" && request.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Nenhum campo para atualizar foi fornecido.",
			"details": "Forneça pelo menos um campo válido (nome ou email).",
		})
		return
	}

	// Construir mapa de dados para atualização
	dados := make(map[string]interface{})
	if request.Nome != "" {
		dados["nome"] = request.Nome
	}
	if request.Email != "" {
		dados["email"] = request.Email
	}

	// Atualizar usuário
	err = h.service.Update(idUrl, empresaID, dados)
	if err != nil {
		errorMsg := err.Error()

		// Classificar erros
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Usuário não encontrado.",
				"details": "O usuário especificado não existe ou não pertence a esta empresa.",
			})
		case errorMsg == "email já está em uso por outro usuário":
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Email já está em uso.",
				"details": "Este endereço de email já está cadastrado para outro usuário.",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Falha ao atualizar o usuário.",
				"details": "Ocorreu um erro ao processar a atualização.",
			})
		}
		return
	}

	// Buscar usuário atualizado para retornar
	usuarioAtualizado, err := h.service.FindByID(idUrl, empresaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Usuário atualizado mas falha ao buscar dados.",
			"details": "O usuário foi atualizado mas não foi possível buscar os dados atualizados.",
		})
		return
	}

	c.JSON(http.StatusOK, usuarioAtualizado)
}

// @Summary      Atualiza parcialmente um usuário (PATCH)
// @Description  Atualiza campos específicos de um usuário (nome, email e/ou cargo). Permite atualização parcial.
// @Description  Regras: ao alterar cargo, valida hierarquia - requisitante não pode atribuir cargo superior ao seu.
// @Tags         Usuários
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                   true  "ID do Usuário a ser atualizado"
// @Param        dados    body      PatchUsuarioRequest   true  "Campos para atualização (parcial)"
// @Success      200      {object}  model.Usuario         "Usuário atualizado com sucesso"
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Router       /usuarios/{id} [patch]
func (h *UsuarioHandler) PatchUsuarioHandler(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Erro de autenticação.",
			"details": "Não foi possível identificar a empresa.",
		})
		return
	}

	idToken, err := h.converter.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Erro de autenticação.",
			"details": "Não foi possível identificar o usuário requisitante.",
		})
		return
	}

	idUrl, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID do usuário inválido.",
			"details": "O ID deve ser um número inteiro positivo.",
		})
		return
	}

	// Verificar permissões
	requester, err := h.service.FindByID(idToken, empresaID)
	if err != nil || requester.Contrato.ID == 0 || requester.Contrato.Cargo.ID == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Acesso negado.",
			"details": "Não foi possível verificar as suas permissões.",
		})
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
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Você não tem permissão para editar este usuário.",
			"details": "Permissões necessárias: EDITAR_USUARIO ou EDITAR_PROPRIA_CONTA.",
		})
		return
	}

	// Parse request (PATCH permite campos opcionais)
	var request PatchUsuarioRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dados de requisição inválidos.",
			"details": err.Error(),
		})
		return
	}

	// Verificar se há pelo menos um campo para atualizar
	if request.Nome == nil && request.Email == nil && request.CargoID == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Nenhum campo para atualizar foi fornecido.",
			"details": "Forneça pelo menos um campo válido (nome, email ou cargoId).",
		})
		return
	}

	// Construir mapa de dados para atualização
	dados := make(map[string]interface{})
	if request.Nome != nil {
		dados["nome"] = *request.Nome
	}
	if request.Email != nil {
		dados["email"] = *request.Email
	}
	if request.CargoID != nil {
		dados["cargo_id"] = *request.CargoID
	}

	// Atualizar usuário (service validará hierarquia se cargo for alterado)
	err = h.service.UpdateWithHierarchy(idUrl, empresaID, idToken, dados)
	if err != nil {
		errorMsg := err.Error()

		// Classificar erros
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Usuário não encontrado.",
				"details": "O usuário especificado não existe ou não pertence a esta empresa.",
			})
		case errorMsg == "email já está em uso por outro usuário":
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Email já está em uso.",
				"details": "Este endereço de email já está cadastrado para outro usuário.",
			})
		case errorMsg == "cargo não encontrado ou não pertence a esta empresa":
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Cargo inválido.",
				"details": "O cargo especificado não existe ou não pertence a esta empresa.",
			})
		case errorMsg == "acesso negado: você não pode atribuir um cargo com nível hierárquico superior ao seu":
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Acesso negado.",
				"details": errorMsg,
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Falha ao atualizar o usuário.",
				"details": errorMsg,
			})
		}
		return
	}

	// Buscar usuário atualizado para retornar
	usuarioAtualizado, err := h.service.FindByID(idUrl, empresaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Usuário atualizado mas falha ao buscar dados.",
			"details": "O usuário foi atualizado mas não foi possível buscar os dados atualizados.",
		})
		return
	}

	c.JSON(http.StatusOK, usuarioAtualizado)
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
// @Failure      404      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Router       /usuarios [post]
func (h *UsuarioHandler) CriarUsuarioHandler(c *gin.Context) {
	var request CriarUsuarioRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dados de requisição inválidos.",
			"details": "Verifique os campos: nome (2-100 caracteres), CPF (11 dígitos), email (válido), senha (mínimo 8 caracteres), salario (maior que 0).",
		})
		return
	}

	// Obter ID do requisitante do contexto
	idRequisitante, err := h.converter.GetUintIDFromContext(c, "userID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Falha na autenticação.",
			"details": "Não foi possível identificar o usuário requisitante.",
		})
		return
	}

	// Força o empresaID a ser o do token para evitar uso indevido ou inconsistências no payload
	empresaIDToken, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Falha na autenticação.",
			"details": "Não foi possível identificar a empresa do requisitante.",
		})
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
		errorMsg := err.Error()

		// Classificar erros e retornar status apropriado
		switch {
		case errorMsg == "email já cadastrado":
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Email já está em uso.",
				"details": "Este endereço de email já está cadastrado no sistema.",
			})
		case errorMsg == "cpf já cadastrado":
			c.JSON(http.StatusConflict, gin.H{
				"error":   "CPF já está em uso.",
				"details": "Este CPF já está cadastrado no sistema.",
			})
		case errorMsg == "cargo não encontrado ou não pertence a esta empresa":
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Cargo não encontrado.",
				"details": "O cargo especificado não existe ou não pertence a esta empresa.",
			})
		case errorMsg == "localidade não encontrada ou não pertence a esta empresa":
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Localidade não encontrada.",
				"details": "A localidade especificada não existe ou não pertence a esta empresa.",
			})
		case errorMsg == "acesso negado: você não pode atribuir um cargo com nível hierárquico superior ao seu":
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Permissão negada.",
				"details": "Você não pode atribuir um cargo com nível hierárquico superior ao seu.",
			})
		case errorMsg == "usuário requisitante não encontrado para validar permissão":
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Falha na validação de permissões.",
				"details": "Não foi possível validar suas permissões.",
			})
		case len(errorMsg) > 20 && errorMsg[:11] == "o salário R":
			// Erro de faixa salarial
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Salário fora da faixa permitida.",
				"details": errorMsg,
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Falha ao criar usuário.",
				"details": "Ocorreu um erro ao processar a criação do usuário.",
			})
		}
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
