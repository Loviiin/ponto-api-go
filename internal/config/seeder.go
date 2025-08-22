package config

import (
	"github.com/Loviiin/ponto-api-go/pkg/password"
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
		{Nome: permissions.EDITAR_USUARIO, Descricao: "Permite editar os dados de outros usuários da empresa."},
		{Nome: permissions.DELETAR_PROPRIA_CONTA, Descricao: "Permite que um usuário delete a sua própria conta."},
		{Nome: permissions.EDITAR_PROPRIA_CONTA, Descricao: "Permite que um usuário edite seus próprios dados."},
		{Nome: permissions.VER_SALDO_FUNCIONARIOS, Descricao: "Permite ver saldo de horas de um funcionário"},
		{Nome: permissions.EDITAR_SALDO_FUNCIONARIOS, Descricao: "Pemite a edição de pontos de um funcionário caso necessário"},
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
		mapaPermissoes[permissions.EDITAR_SALDO_FUNCIONARIOS],
		mapaPermissoes[permissions.VER_SALDO_FUNCIONARIOS],
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

func SeedSuperAdmin(db *gorm.DB) {
	var usuarioExistente model.Usuario
	err := db.Where("email = ?", "superadmin@ponto.com").First(&usuarioExistente).Error
	if err == nil {
		log.Println("Usuário Super Admin já existe.")
		return
	}

	var empresa model.Empresa
	err = db.First(&empresa).Error
	if err != nil {
		empresa = model.Empresa{Nome: "Empresa Padrão"}
		db.Create(&empresa)
	}

	mapaPermissoes := SeedPermissions(db)
	SetupDefaultRolesAndPermissions(db, empresa.ID, mapaPermissoes)

	var adminCargo model.Cargo
	db.Where("nome = ? AND empresa_id = ?", "Admin", empresa.ID).First(&adminCargo)
	if adminCargo.ID == 0 {
		log.Println("Falha ao encontrar o cargo de Admin para o Super Admin.")
		return
	}

	senhaCripto, _ := password.CriptografaSenha("superadmin")
	superAdmin := model.Usuario{
		Nome:      "Super Admin",
		Email:     "superadmin@ponto.com",
		Senha:     string(senhaCripto),
		EmpresaID: empresa.ID,
		CargoID:   adminCargo.ID,
	}

	db.Create(&superAdmin)
	log.Println("Usuário Super Admin criado com sucesso.")
}
