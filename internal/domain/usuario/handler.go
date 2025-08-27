package usuario

import (
	"errors"
	"github.com/Loviiin/ponto-api-go/internal/domain/cargo"
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/Loviiin/ponto-api-go/pkg/permissions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type UsuarioHandler struct {
	service        UsuarioService
	empresaService empresa.EmpresaService
	cargoService   cargo.CargoService
	converter      funcoes.FuncoesInterface
}

func NewUsuarioHandler(s UsuarioService, e empresa.EmpresaService, ca cargo.CargoService, f funcoes.FuncoesInterface) *UsuarioHandler {
	return &UsuarioHandler{
		service:        s,
		empresaService: e,
		cargoService:   ca,
		converter:      f,
	}
}

type CriarUsuarioRequest struct {
	Nome      string `json:"nome" binding:"required" example:"João Silva"`
	Email     string `json:"email" binding:"required,email" example:"joao.silva@empresa.com"`
	Senha     string `json:"senha" binding:"required,min=6" example:"senha123"`
	EmpresaID uint   `json:"empresa_id" binding:"required" example:"1"`
	CargoID   uint   `json:"cargo_id" binding:"required" example:"2"`
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
// @Success      200  {array}   model.Usuario
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
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado."})
		return
	}

	podeDeletar := false
	if idUrl == idToken {
		for _, p := range requester.Cargo.Permissoes {
			if p.Nome == permissions.DELETAR_PROPRIA_CONTA {
				podeDeletar = true
				break
			}
		}
	} else {
		for _, p := range requester.Cargo.Permissoes {
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
// @Param        dados    body      UpdateUsuarioRequest true  "Dados para atualização" // <-- Usamos a struct aqui para o Swagger
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
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado."})
		return
	}

	podeEditar := false
	if idUrl == idToken {
		for _, p := range requester.Cargo.Permissoes {
			if p.Nome == permissions.EDITAR_PROPRIA_CONTA {
				podeEditar = true
				break
			}
		}
	} else {
		for _, p := range requester.Cargo.Permissoes {
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

	if idUrl == idToken {
		delete(dadosParaAtualizar, "cargo_id")
	}
	delete(dadosParaAtualizar, "empresa_id")

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
// @Description  Cria um novo usuário (funcionário) no sistema.
// @Tags         Usuários
// @Accept       json
// @Produce      json
// @Param        usuario  body      CriarUsuarioRequest  true  "Dados do Novo Usuário"
// @Success      201      {object}  model.Usuario
// @Failure      400      {object}  map[string]string
// @Router       /usuarios [post]
func (h *UsuarioHandler) CriarUsuarioHandler(c *gin.Context) {

	var request CriarUsuarioRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.empresaService.GetEmpresaByIDSer(request.EmpresaID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A empresa especificada não existe."})
		return
	}

	_, err = h.cargoService.FindByID(request.CargoID, request.EmpresaID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O cargo especificado não existe ou não pertence a esta empresa."})
		return
	}
	usuario := model.Usuario{
		Nome:      request.Nome,
		Email:     request.Email,
		Senha:     request.Senha,
		EmpresaID: request.EmpresaID,
		CargoID:   request.CargoID,
	}

	err = h.service.CriarUsuario(&usuario) // Este método agora será mais simples!
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
