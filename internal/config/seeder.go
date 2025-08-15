package config

import (
	"log"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/permissions"
	"gorm.io/gorm"
)

// SeedPermissions cria as permissões padrão no sistema se elas não existirem.
func SeedPermissions(db *gorm.DB) map[string]model.Permissao {
	permissoes := []model.Permissao{
		{Nome: permissions.EDITAR_EMPRESA, Descricao: "Permite editar os dados da própria empresa."},
		{Nome: permissions.DELETAR_EMPRESA, Descricao: "Permite deletar a própria empresa."},
		{Nome: permissions.GERENCIAR_CARGOS, Descricao: "Permite criar, editar, apagar e gerenciar permissões de cargos."},
		{Nome: permissions.DELETAR_USUARIO, Descricao: "Permite deletar outros usuários da empresa."},
		{Nome: permissions.EDITAR_USUARIO, Descricao: "Permite editar os dados de outros usuários da empresa."}, // <-- ADICIONADA
		{Nome: permissions.DELETAR_PROPRIA_CONTA, Descricao: "Permite que um usuário delete a sua própria conta."},
		{Nome: permissions.EDITAR_PROPRIA_CONTA, Descricao: "Permite que um usuário edite seus próprios dados."},
	}

	for i := range permissoes {
		db.FirstOrCreate(&permissoes[i], model.Permissao{Nome: permissoes[i].Nome})
	}
	log.Println("Permissões padrão verificadas/criadas.")

	mapaPermissoes := make(map[string]model.Permissao)
	for _, p := range permissoes {
		mapaPermissoes[p.Nome] = p
	}
	return mapaPermissoes
}

func SetupDefaultRolesAndPermissions(db *gorm.DB, empresaID uint, mapaPermissoes map[string]model.Permissao) {
	adminRole := model.Cargo{Nome: "Admin", EmpresaID: empresaID}
	db.Where(model.Cargo{Nome: adminRole.Nome, EmpresaID: empresaID}).FirstOrCreate(&adminRole)

	funcRole := model.Cargo{Nome: "Funcionário", EmpresaID: empresaID}
	db.Where(model.Cargo{Nome: funcRole.Nome, EmpresaID: empresaID}).FirstOrCreate(&funcRole)

	adminPermissions := []model.Permissao{
		mapaPermissoes[permissions.EDITAR_EMPRESA],
		mapaPermissoes[permissions.DELETAR_EMPRESA],
		mapaPermissoes[permissions.GERENCIAR_CARGOS],
		mapaPermissoes[permissions.DELETAR_USUARIO],
		mapaPermissoes[permissions.EDITAR_USUARIO],
		mapaPermissoes[permissions.DELETAR_PROPRIA_CONTA],
		mapaPermissoes[permissions.EDITAR_PROPRIA_CONTA],
	}

	funcPermissions := []model.Permissao{
		mapaPermissoes[permissions.DELETAR_PROPRIA_CONTA],
		mapaPermissoes[permissions.EDITAR_PROPRIA_CONTA],
	}

	err := db.Model(&adminRole).Association("Permissoes").Replace(adminPermissions)
	if err != nil {
		return
	}
	err = db.Model(&funcRole).Association("Permissoes").Replace(funcPermissions)
	if err != nil {
		return
	}

	log.Printf("Cargos e permissões padrão configurados para a empresa %d.", empresaID)
}
