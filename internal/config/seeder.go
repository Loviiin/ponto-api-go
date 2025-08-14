package config

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
	"log"
)

// SeedPermissions cria as permissões padrão no sistema se elas não existirem.
func SeedPermissions(db *gorm.DB) map[string]model.Permissao {
	permissoes := []model.Permissao{
		{Nome: "EDITAR_EMPRESA", Descricao: "Permite editar os dados da própria empresa."},
		{Nome: "DELETAR_EMPRESA", Descricao: "Permite deletar a própria empresa."},
		{Nome: "DELETAR_USUARIO", Descricao: "Permite deletar outros usuários da empresa."},
		{Nome: "GERENCIAR_CARGOS", Descricao: "Permite criar, editar, apagar e gerenciar permissões de cargos."},
		{Nome: "DELETAR_PROPRIA_CONTA", Descricao: "Permite que um usuário delete a sua própria conta."},
	}

	for i := range permissoes {
		// FirstOrCreate encontra a permissão ou a cria se não existir.
		db.FirstOrCreate(&permissoes[i], model.Permissao{Nome: permissoes[i].Nome})
	}
	log.Println("Permissões padrão verificadas/criadas.")

	// Retornamos um mapa para fácil acesso no futuro
	mapaPermissoes := make(map[string]model.Permissao)
	for _, p := range permissoes {
		mapaPermissoes[p.Nome] = p
	}
	return mapaPermissoes
}

// SetupDefaultRolesAndPermissions cria os cargos padrão para uma nova empresa e associa as permissões corretas.
func SetupDefaultRolesAndPermissions(db *gorm.DB, empresaID uint, mapaPermissoes map[string]model.Permissao) {
	// --- Cargo de Admin ---
	adminRole := model.Cargo{
		Nome:      "Admin",
		EmpresaID: empresaID,
	}
	// Usamos FirstOrCreate para evitar duplicados se a função for chamada mais de uma vez.
	if err := db.Where(model.Cargo{Nome: adminRole.Nome, EmpresaID: empresaID}).FirstOrCreate(&adminRole).Error; err != nil {
		log.Printf("Erro ao criar o cargo Admin para a empresa %d: %v", empresaID, err)
		return
	}

	// --- Cargo de Funcionário ---
	funcRole := model.Cargo{
		Nome:      "Funcionário",
		EmpresaID: empresaID,
	}
	if err := db.Where(model.Cargo{Nome: funcRole.Nome, EmpresaID: empresaID}).FirstOrCreate(&funcRole).Error; err != nil {
		log.Printf("Erro ao criar o cargo Funcionário para a empresa %d: %v", empresaID, err)
		return
	}

	// --- Associar Permissões ---
	// O Admin tem todas as permissões.
	adminPermissions := []model.Permissao{
		mapaPermissoes["EDITAR_EMPRESA"],
		mapaPermissoes["DELETAR_EMPRESA"],
		mapaPermissoes["DELETAR_USUARIO"],
		mapaPermissoes["GERENCIAR_CARGOS"],
		mapaPermissoes["DELETAR_PROPRIA_CONTA"],
	}

	// O Funcionário só pode apagar a própria conta.
	funcPermissions := []model.Permissao{
		mapaPermissoes["DELETAR_PROPRIA_CONTA"],
	}

	db.Model(&adminRole).Association("Permissoes").Replace(adminPermissions)
	db.Model(&funcRole).Association("Permissoes").Replace(funcPermissions)

	log.Printf("Cargos e permissões padrão configurados para a empresa %d.", empresaID)
}
