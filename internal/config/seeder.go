package config

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Loviiin/ponto-api-go/pkg/password"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/permissions"
	"gorm.io/gorm"
)

const (
	demoEmpresaCNPJ      = "00000000000192"
	demoEmpresaNome      = "Demo Ponto"
	demoLocalidadeNome   = "Matriz Demo"
	demoLocalidadeCEP    = "01001-000"
	demoUsuarioNome      = "Demo User"
	demoUsuarioEmail     = "demo@ponto.com"
	demoUsuarioCPF       = "11144477735"
	demoUsuarioSenha     = "Demo@12345"
	demoUsuarioSalario   = 0.0
	demoCargoNome        = "Dono"
	demoCargoHierarquia  = 100
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
		{Nome: permissions.GERENCIAR_LOCALIDADES, Descricao: "Permite criar, editar e apagar localidades da empresa."},
		{Nome: permissions.VER_JUSTIFICATIVAS_PENDENTES, Descricao: "Permite visualizar justificativas pendentes de aprovação."},
		{Nome: permissions.CRIAR_JUSTIFICATIVA_PROPRIA, Descricao: "Permite que o funcionário crie justificativas para seus próprios pontos."},
		{Nome: permissions.APROVAR_JUSTIFICATIVAS, Descricao: "Permite aprovar ou reprovar justificativas de ponto dos funcionários."},
		{Nome: permissions.VISUALIZAR_RELATORIOS_GERAIS, Descricao: "Permite visualizar e exportar relatórios gerais de ponto de todos os funcionários."},
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
	dono = model.Cargo{Nome: "Dono", EmpresaID: empresaID, NivelHierarquia: 100}
	db.Where(model.Cargo{Nome: dono.Nome, EmpresaID: empresaID}).FirstOrCreate(&dono)

	gerente = model.Cargo{Nome: "Gerente", EmpresaID: empresaID, NivelHierarquia: 70}
	db.Where(model.Cargo{Nome: gerente.Nome, EmpresaID: empresaID}).FirstOrCreate(&gerente)

	colaborador = model.Cargo{Nome: "Colaborador", EmpresaID: empresaID, NivelHierarquia: 10}
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
		mapaPermissoes[permissions.GERENCIAR_LOCALIDADES],
		mapaPermissoes[permissions.VER_JUSTIFICATIVAS_PENDENTES],
		mapaPermissoes[permissions.CRIAR_JUSTIFICATIVA_PROPRIA],
		mapaPermissoes[permissions.APROVAR_JUSTIFICATIVAS],
		mapaPermissoes[permissions.VISUALIZAR_RELATORIOS_GERAIS],
	}
	gerentePerms := []model.Permissao{
		mapaPermissoes[permissions.GERENCIAR_CARGOS],
		mapaPermissoes[permissions.EDITAR_USUARIO],
		mapaPermissoes[permissions.DELETAR_USUARIO],
		mapaPermissoes[permissions.EDITAR_PROPRIA_CONTA],
		mapaPermissoes[permissions.DELETAR_PROPRIA_CONTA],
		mapaPermissoes[permissions.EDITAR_SALDO_FUNCIONARIOS],
		mapaPermissoes[permissions.VER_SALDO_FUNCIONARIOS],
		mapaPermissoes[permissions.VISUALIZAR_PONTO_FUNCIONARIOS],
		mapaPermissoes[permissions.AJUSTAR_PONTO_FUNCIONARIOS],
		mapaPermissoes[permissions.GERENCIAR_JUSTIFICATIVAS],
		mapaPermissoes[permissions.VER_JUSTIFICATIVAS_PENDENTES],
		mapaPermissoes[permissions.CRIAR_JUSTIFICATIVA_PROPRIA],
		mapaPermissoes[permissions.APROVAR_JUSTIFICATIVAS],
		// Correção: permitir que Gerente gerencie localidades (necessário para cadastro)
		mapaPermissoes[permissions.GERENCIAR_LOCALIDADES],
		mapaPermissoes[permissions.VISUALIZAR_RELATORIOS_GERAIS],
	}
	colaboradorPerms := []model.Permissao{
		mapaPermissoes[permissions.EDITAR_PROPRIA_CONTA],
		mapaPermissoes[permissions.DELETAR_PROPRIA_CONTA],
		mapaPermissoes[permissions.CRIAR_JUSTIFICATIVA_PROPRIA],
	}

	db.Model(&dono).Association("Permissoes").Replace(donoPerms)
	db.Model(&gerente).Association("Permissoes").Replace(gerentePerms)
	db.Model(&colaborador).Association("Permissoes").Replace(colaboradorPerms)

	log.Printf("Cargos e permissões padrão configurados para a empresa %d.", empresaID)
	return
}

func SeedSuperAdmin(db *gorm.DB) {
	// Garantir empresa e localidade padrão
	empresa := model.Empresa{
		NomeFantasia: "Empresa Padrão",
		RazaoSocial:  "Empresa Padrão LTDA",
		CNPJ:         "00000000000191",
	}
	db.Where(model.Empresa{CNPJ: empresa.CNPJ}).FirstOrCreate(&empresa)

	localidade := model.Localidade{
		Nome:               "Matriz Padrão",
		EmpresaID:          empresa.ID,
		CEP:                "01001-000",
		Cidade:             "São Paulo",
		Estado:             "SP",
		Latitude:           -23.550520,
		Longitude:          -46.633308,
		RaioGeofenceMetros: 100,
	}
	db.Where(model.Localidade{EmpresaID: empresa.ID, Nome: "Matriz Padrão"}).FirstOrCreate(&localidade)

	// Permissões e cargos padrão
	mapaPermissoes := SeedPermissions(db)
	donoCargo, gerenteCargo, colaboradorCargo := SetupDefaultRolesAndPermissions(db, empresa.ID, mapaPermissoes)

	// Helper para criar usuário e contrato se necessário
	createUserWithContract := func(nome, email, cpf, senha string, cargoID uint) {
		// Usuário
		var u model.Usuario
		err := db.Where("email = ?", email).First(&u).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			hash, _ := password.CriptografaSenha(senha)
			u = model.Usuario{
				Nome:  nome,
				Email: email,
				CPF:   cpf,
				Senha: hash,
			}
			if e := db.Create(&u).Error; e != nil {
				log.Printf("Falha ao criar usuário %s: %v", email, e)
				return
			}
		} else if err != nil {
			log.Printf("Erro ao buscar usuário %s: %v", email, err)
			return
		}

		// Contrato
		var c model.Contrato
		err = db.Where("usuario_id = ?", u.ID).First(&c).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c = model.Contrato{
				UsuarioID:    u.ID,
				EmpresaID:    empresa.ID,
				LocalidadeID: localidade.ID,
				CargoID:      cargoID,
				DataAdmissao: time.Now(),
				Salario:      0,
			}
			if e := db.Create(&c).Error; e != nil {
				log.Printf("Falha ao criar contrato para %s: %v", email, e)
				return
			}
		} else if err != nil {
			log.Printf("Erro ao buscar contrato do usuário %s: %v", email, err)
			return
		}
	}

	// Usuários padrão
	createUserWithContract("Dono", "dono@ponto.com", "00000000001", "donosenha", donoCargo.ID)
	createUserWithContract("Gerente", "gerente@ponto.com", "00000000002", "gerentesenha", gerenteCargo.ID)
	createUserWithContract("Colaborador", "colaborador@ponto.com", "00000000003", "colabsenha", colaboradorCargo.ID)

	// Super Admin global (para desenvolvimento)
	var super model.Usuario
	err := db.Where("email = ?", "superadmin@ponto.com").First(&super).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hash, _ := password.CriptografaSenha("superadmin")
		super = model.Usuario{
			Nome:  "Super Admin",
			Email: "superadmin@ponto.com",
			CPF:   "00000000000",
			Senha: hash,
		}
		if e := db.Create(&super).Error; e != nil {
			log.Printf("Falha ao criar Super Admin: %v", e)
			return
		}

		superCargo := model.Cargo{
			Nome:            "Super Admin",
			EmpresaID:       empresa.ID,
			NivelHierarquia: 1000000,
		}

		// Criar o cargo primeiro
		db.Where(model.Cargo{Nome: superCargo.Nome, EmpresaID: empresa.ID}).FirstOrCreate(&superCargo)

		// Depois associar todas as permissões ao Super Admin
		var todasPermissoes []model.Permissao
		for _, p := range mapaPermissoes {
			todasPermissoes = append(todasPermissoes, p)
		}
		db.Model(&superCargo).Association("Permissoes").Replace(todasPermissoes)

		c := model.Contrato{
			UsuarioID:    super.ID,
			EmpresaID:    empresa.ID,
			LocalidadeID: localidade.ID,
			CargoID:      superCargo.ID,
			DataAdmissao: time.Now(),
			Salario:      99999.0,
		}
		if e := db.Create(&c).Error; e != nil {
			log.Printf("Falha ao criar contrato do Super Admin: %v", e)
			return
		}
		log.Println("Usuário Super Admin criado com sucesso.")
	} else if err != nil {
		log.Printf("Erro ao buscar Super Admin: %v", err)
	}
}

// ResetAndSeedDemoWorkspace recria o tenant demo do portfólio com credenciais fixas.
func ResetAndSeedDemoWorkspace(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := resetDemoWorkspace(tx); err != nil {
			return err
		}
		return seedDemoWorkspace(tx)
	})
}

func seedDemoWorkspace(db *gorm.DB) error {
	empresa := model.Empresa{
		NomeFantasia: demoEmpresaNome,
		RazaoSocial:  fmt.Sprintf("%s LTDA", demoEmpresaNome),
		CNPJ:         demoEmpresaCNPJ,
	}
	if err := db.Where(model.Empresa{CNPJ: demoEmpresaCNPJ}).FirstOrCreate(&empresa).Error; err != nil {
		return err
	}

	localidade := model.Localidade{
		Nome:               demoLocalidadeNome,
		EmpresaID:          empresa.ID,
		CEP:                demoLocalidadeCEP,
		Cidade:             "São Paulo",
		Estado:             "SP",
		Latitude:           -23.550520,
		Longitude:          -46.633308,
		RaioGeofenceMetros: 100,
	}
	if err := db.Where(model.Localidade{EmpresaID: empresa.ID, Nome: demoLocalidadeNome}).FirstOrCreate(&localidade).Error; err != nil {
		return err
	}

	mapaPermissoes := SeedPermissions(db)
	donoCargo, _, _ := SetupDefaultRolesAndPermissions(db, empresa.ID, mapaPermissoes)
	if donoCargo.ID == 0 {
		return fmt.Errorf("falha ao preparar cargo demo")
	}

	var usuarioDemo model.Usuario
	err := db.Where("email = ?", demoUsuarioEmail).First(&usuarioDemo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hash, hashErr := password.CriptografaSenha(demoUsuarioSenha)
		if hashErr != nil {
			return hashErr
		}
		usuarioDemo = model.Usuario{
			Nome:  demoUsuarioNome,
			Email: demoUsuarioEmail,
			CPF:   demoUsuarioCPF,
			Senha: hash,
		}
		if err := db.Create(&usuarioDemo).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	var contratoDemo model.Contrato
	err = db.Where("usuario_id = ?", usuarioDemo.ID).First(&contratoDemo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		contratoDemo = model.Contrato{
			UsuarioID:    usuarioDemo.ID,
			EmpresaID:    empresa.ID,
			LocalidadeID: localidade.ID,
			CargoID:      donoCargo.ID,
			Salario:      demoUsuarioSalario,
			DataAdmissao: time.Now(),
		}
		if err := db.Create(&contratoDemo).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	log.Printf("Demo workspace pronto: %s / %s", demoUsuarioEmail, demoUsuarioSenha)
	return nil
}

func resetDemoWorkspace(db *gorm.DB) error {
	var empresa model.Empresa
	if err := db.Where("cnpj = ?", demoEmpresaCNPJ).First(&empresa).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	var cargos []model.Cargo
	if err := db.Where("empresa_id = ?", empresa.ID).Find(&cargos).Error; err != nil {
		return err
	}
	for i := range cargos {
		if err := db.Model(&cargos[i]).Association("Permissoes").Clear(); err != nil {
			return err
		}
	}

	var contratos []model.Contrato
	if err := db.Where("empresa_id = ?", empresa.ID).Find(&contratos).Error; err != nil {
		return err
	}

	var usuarioIDs []uint
	for _, contrato := range contratos {
		usuarioIDs = append(usuarioIDs, contrato.UsuarioID)
	}

	if err := db.Where("empresa_id = ?", empresa.ID).Delete(&model.Justificativa{}).Error; err != nil {
		return err
	}
	if err := db.Where("empresa_id = ?", empresa.ID).Delete(&model.LogBancoHoras{}).Error; err != nil {
		return err
	}
	if err := db.Where("empresa_id = ?", empresa.ID).Delete(&model.RegistroPonto{}).Error; err != nil {
		return err
	}
	if err := db.Where("empresa_id = ?", empresa.ID).Delete(&model.Contrato{}).Error; err != nil {
		return err
	}
	if err := db.Where("empresa_id = ?", empresa.ID).Delete(&model.AuditLog{}).Error; err != nil {
		return err
	}

	if len(usuarioIDs) > 0 {
		if err := db.Where("usuario_id IN ?", usuarioIDs).Delete(&model.PasswordResetToken{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("id IN ?", usuarioIDs).Delete(&model.Usuario{}).Error; err != nil {
			return err
		}
	}

	if err := db.Unscoped().Where("empresa_id = ?", empresa.ID).Delete(&model.Cargo{}).Error; err != nil {
		return err
	}
	if err := db.Where("empresa_id = ?", empresa.ID).Delete(&model.Localidade{}).Error; err != nil {
		return err
	}
	if err := db.Delete(&empresa).Error; err != nil {
		return err
	}

	return nil
}
