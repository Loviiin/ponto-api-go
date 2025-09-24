package config

import (
	"log"

	"github.com/Loviiin/ponto-api-go/pkg/password"

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
		{Nome: permissions.VISUALIZAR_PONTO_FUNCIONARIOS, Descricao: "Permite visualizar os registros de ponto de outros funcionários."},
		{Nome: permissions.AJUSTAR_PONTO_FUNCIONARIOS, Descricao: "Permite adicionar, editar ou remover registros de ponto de outros funcionários."},
		{Nome: permissions.GERENCIAR_JUSTIFICATIVAS, Descricao: "Permite gerenciar justificativas de ponto dos funcionários."},
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

func SetupDefaultRolesAndPermissions(db *gorm.DB, empresaID uint, mapaPermissoes map[string]model.Permissao) (dono model.Cargo, gerente model.Cargo, colaborador model.Cargo) {
	// Criação dos cargos
	dono = model.Cargo{Nome: "Dono", EmpresaID: empresaID}
	db.Where(model.Cargo{Nome: dono.Nome, EmpresaID: empresaID}).FirstOrCreate(&dono)

	gerente = model.Cargo{Nome: "Gerente", EmpresaID: empresaID}
	db.Where(model.Cargo{Nome: gerente.Nome, EmpresaID: empresaID}).FirstOrCreate(&gerente)

	colaborador = model.Cargo{Nome: "Colaborador", EmpresaID: empresaID}
	db.Where(model.Cargo{Nome: colaborador.Nome, EmpresaID: empresaID}).FirstOrCreate(&colaborador)

	// Permissões para cada cargo
	donoPerms := []model.Permissao{
		mapaPermissoes[permissions.EDITAR_EMPRESA],
		mapaPermissoes[permissions.DELETAR_EMPRESA],
		mapaPermissoes[permissions.GERENCIAR_CARGOS],
		mapaPermissoes[permissions.DELETAR_USUARIO],
		mapaPermissoes[permissions.EDITAR_USUARIO],
		mapaPermissoes[permissions.DELETAR_PROPRIA_CONTA],
		mapaPermissoes[permissions.EDITAR_PROPRIA_CONTA],
		mapaPermissoes[permissions.EDITAR_SALDO_FUNCIONARIOS],
		mapaPermissoes[permissions.VER_SALDO_FUNCIONARIOS],
		mapaPermissoes[permissions.VISUALIZAR_PONTO_FUNCIONARIOS],
		mapaPermissoes[permissions.AJUSTAR_PONTO_FUNCIONARIOS],
		mapaPermissoes[permissions.GERENCIAR_JUSTIFICATIVAS],
	}
	gerentePerms := []model.Permissao{
		mapaPermissoes[permissions.GERENCIAR_CARGOS],
		mapaPermissoes[permissions.EDITAR_USUARIO],
		mapaPermissoes[permissions.EDITAR_PROPRIA_CONTA],
		mapaPermissoes[permissions.DELETAR_PROPRIA_CONTA],
		mapaPermissoes[permissions.EDITAR_SALDO_FUNCIONARIOS],
		mapaPermissoes[permissions.VER_SALDO_FUNCIONARIOS],
		mapaPermissoes[permissions.VISUALIZAR_PONTO_FUNCIONARIOS],
		mapaPermissoes[permissions.AJUSTAR_PONTO_FUNCIONARIOS],
		mapaPermissoes[permissions.GERENCIAR_JUSTIFICATIVAS],
	}
	colaboradorPerms := []model.Permissao{
		mapaPermissoes[permissions.EDITAR_PROPRIA_CONTA],
		mapaPermissoes[permissions.DELETAR_PROPRIA_CONTA],
	}

	db.Model(&dono).Association("Permissoes").Replace(donoPerms)
	db.Model(&gerente).Association("Permissoes").Replace(gerentePerms)
	db.Model(&colaborador).Association("Permissoes").Replace(colaboradorPerms)

	log.Printf("Cargos e permissões padrão configurados para a empresa %d.", empresaID)
	return
}

func SeedSuperAdmin(db *gorm.DB) {
	// Criação dos usuários padrão da empresa
	var usuarioExistente model.Usuario
	emails := []string{"dono@ponto.com", "gerente@ponto.com", "colaborador@ponto.com"}
	for _, email := range emails {
		err := db.Where("email = ?", email).First(&usuarioExistente).Error
		if err == nil {
			log.Printf("Usuário %s já existe.", email)
			return
		}
	}

	var empresa model.Empresa
	err := db.First(&empresa).Error
	if err != nil {
		empresa = model.Empresa{Nome: "Empresa Padrão"}
		db.Create(&empresa)
	}

	mapaPermissoes := SeedPermissions(db)
	donoCargo, gerenteCargo, colaboradorCargo := SetupDefaultRolesAndPermissions(db, empresa.ID, mapaPermissoes)

	// Criação dos usuários
	senhaDono, _ := password.CriptografaSenha("donosenha")
	dono := model.Usuario{
		Nome:      "Dono",
		Email:     "dono@ponto.com",
		Senha:     string(senhaDono),
		EmpresaID: empresa.ID,
		CargoID:   donoCargo.ID,
	}
	senhaGerente, _ := password.CriptografaSenha("gerentesenha")
	gerente := model.Usuario{
		Nome:      "Gerente",
		Email:     "gerente@ponto.com",
		Senha:     string(senhaGerente),
		EmpresaID: empresa.ID,
		CargoID:   gerenteCargo.ID,
	}
	senhaColab, _ := password.CriptografaSenha("colabsenha")
	colaborador := model.Usuario{
		Nome:      "Colaborador",
		Email:     "colaborador@ponto.com",
		Senha:     string(senhaColab),
		EmpresaID: empresa.ID,
		CargoID:   colaboradorCargo.ID,
	}

	db.Create(&dono)
	db.Create(&gerente)
	db.Create(&colaborador)
	log.Println("Usuários Dono, Gerente e Colaborador criados com sucesso.")

	// Criação do Super Admin global (para desenvolvedores)
	var superAdminExistente model.Usuario
	err = db.Where("email = ?", "superadmin@ponto.com").First(&superAdminExistente).Error
	if err == nil {
		log.Println("Super Admin já existe.")
		return
	}

	// Vincular Super Admin à empresa padrão e ao cargo Dono
	senhaSuper, _ := password.CriptografaSenha("superadmin")
	superAdmin := model.Usuario{
		Nome:      "Super Admin",
		Email:     "superadmin@ponto.com",
		Senha:     string(senhaSuper),
		EmpresaID: empresa.ID,
		CargoID:   donoCargo.ID,
	}
	db.Create(&superAdmin)
	log.Println("Super Admin criado com sucesso.")
}
