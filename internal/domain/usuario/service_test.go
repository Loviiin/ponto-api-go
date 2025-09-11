package usuario

import (
	"errors"
	"testing"

	"github.com/Loviiin/ponto-api-go/internal/domain/cargo"
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

// Métodos dummy para implementar a interface completa de UsuarioRepository
func (m *mockUsuarioRepository) WithTransaction(tx *gorm.DB) UsuarioRepository { return m }

// Métodos dummy para implementar a interface completa de CargoRepository
func (m *mockCargoRepository) Create(cargo *model.Cargo) error { return nil }
func (m *mockCargoRepository) GetAllByEmpresaID(empresaID uint) ([]model.Cargo, error) {
	return nil, nil
}
func (m *mockCargoRepository) Update(id uint, empresaID uint, dados map[string]interface{}) error {
	return nil
}
func (m *mockCargoRepository) Delete(id uint, empresaID uint) error                      { return nil }
func (m *mockCargoRepository) AddPermissionToCargo(cargoID uint, permissaoID uint) error { return nil }
func (m *mockCargoRepository) WithTransaction(tx *gorm.DB) cargo.CargoRepository         { return m }

// Métodos dummy para implementar a interface completa de EmpresaRepository
func (m *mockEmpresaRepository) CreateEmpresa(empresa *model.Empresa) error { return nil }
func (m *mockEmpresaRepository) GetAllEmpresas() ([]model.Empresa, error)   { return nil, nil }
func (m *mockEmpresaRepository) GetEmpresaByID(idempresa uint) (*model.Empresa, error) {
	return nil, nil
}
func (m *mockEmpresaRepository) UpdateEmpresa(idempresa uint, dados map[string]interface{}) error {
	return nil
}
func (m *mockEmpresaRepository) DeleteEmpresa(idempresa uint) error                    { return nil }
func (m *mockEmpresaRepository) WithTransaction(tx *gorm.DB) empresa.EmpresaRepository { return m }

type mockUsuarioRepository struct {
	SaveFunc        func(usuario *model.Usuario) error
	FindByEmailFunc func(email string) (*model.Usuario, error)
	FindByIDFunc    func(id uint, empresaID uint) (*model.Usuario, error)
	GetAllFunc      func(empresaID uint) ([]model.Usuario, error)
	UpdateFunc      func(id uint, empresaID uint, dados map[string]interface{}) error
	DeleteFunc      func(id uint, empresaID uint) error
	FindAllFunc     func() ([]model.Usuario, error)
}

func (m *mockUsuarioRepository) Save(usuario *model.Usuario) error {
	return m.SaveFunc(usuario)
}
func (m *mockUsuarioRepository) FindByEmail(email string) (*model.Usuario, error) {
	return m.FindByEmailFunc(email)
}
func (m *mockUsuarioRepository) FindByID(id uint, empresaID uint) (*model.Usuario, error) {
	return m.FindByIDFunc(id, empresaID)
}
func (m *mockUsuarioRepository) GetAll(empresaID uint) ([]model.Usuario, error) {
	return m.GetAllFunc(empresaID)
}
func (m *mockUsuarioRepository) Update(id uint, empresaID uint, dados map[string]interface{}) error {
	return m.UpdateFunc(id, empresaID, dados)
}
func (m *mockUsuarioRepository) Delete(id uint, empresaID uint) error {
	return m.DeleteFunc(id, empresaID)
}
func (m *mockUsuarioRepository) FindAll() ([]model.Usuario, error) {
	return m.FindAllFunc()
}

type mockCargoRepository struct {
	FindByIDFunc   func(id uint, empresaID uint) (*model.Cargo, error)
	FindByNameFunc func(nome string, empresaID uint) (*model.Cargo, error)
}

func (m *mockCargoRepository) FindByID(id uint, empresaID uint) (*model.Cargo, error) {
	return m.FindByIDFunc(id, empresaID)
}
func (m *mockCargoRepository) FindByName(nome string, empresaID uint) (*model.Cargo, error) {
	return m.FindByNameFunc(nome, empresaID)
}

type mockEmpresaRepository struct {
	FindByIDFunc func(id uint) (*model.Empresa, error)
}

func (m *mockEmpresaRepository) FindByID(id uint) (*model.Empresa, error) {
	return m.FindByIDFunc(id)
}

func TestCriarUsuario_ComSucesso(t *testing.T) {

	mockRepo := &mockUsuarioRepository{}
	mockCargoRepo := &mockCargoRepository{}
	mockEmpresaRepo := &mockEmpresaRepository{}
	service := NewUsuarioService(mockRepo, mockCargoRepo, mockEmpresaRepo)

	usuarioParaCriar := &model.Usuario{
		Nome:      "Usuário de Teste",
		Email:     "sucesso@email.com",
		Senha:     "senha123",
		EmpresaID: 1,
		CargoID:   2,
	}

	mockRepo.FindByEmailFunc = func(email string) (*model.Usuario, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockEmpresaRepo.FindByIDFunc = func(id uint) (*model.Empresa, error) {
		return &model.Empresa{ID: id, Nome: "Empresa Teste"}, nil
	}
	mockCargoRepo.FindByIDFunc = func(id uint, empresaID uint) (*model.Cargo, error) {
		return &model.Cargo{ID: id, EmpresaID: empresaID, Nome: "Testador"}, nil
	}
	mockRepo.SaveFunc = func(usuario *model.Usuario) error {
		return nil
	}

	err := service.CriarUsuario(usuarioParaCriar, "")
	if err != nil {
		t.Errorf("Erro inesperado ao criar usuário: %v", err)
	}
}

func TestFindByIDComSucesso(t *testing.T) {

	mockRepo := &mockUsuarioRepository{}
	mockCargoRepo := &mockCargoRepository{}
	mockEmpresaRepo := &mockEmpresaRepository{}
	service := NewUsuarioService(mockRepo, mockCargoRepo, mockEmpresaRepo)

	usuarioID := uint(1)
	empresaID := uint(1)
	usuarioEsperado := &model.Usuario{
		ID:        usuarioID,
		Nome:      "Usuário de Teste",
		Email:     "sucesso@email.com",
		Senha:     "senha123",
		EmpresaID: empresaID,
	}

	mockRepo.FindByIDFunc = func(id uint, empID uint) (*model.Usuario, error) {
		if id == usuarioID && empID == empresaID {
			return usuarioEsperado, nil
		}
		return nil, gorm.ErrRecordNotFound
	}

	usuario, err := service.FindByID(usuarioID, empresaID)

	if err != nil {
		t.Fatalf("Esperava não ter erro, mas recebeu: %v", err)
	}
	if usuario == nil {
		t.Fatal("Esperava um usuário, mas recebeu nil")
	}
	if usuario.ID != usuarioEsperado.ID {
		t.Errorf("ID incorreto. Esperava '%d', mas recebeu '%d'", usuarioEsperado.ID, usuario.ID)
	}
	if usuario.Nome != usuarioEsperado.Nome {
		t.Errorf("Nome incorreto. Esperava '%s', mas recebeu '%s'", usuarioEsperado.Nome, usuario.Nome)
	}
}

func TestCriarUsuario_EmailJaExiste(t *testing.T) {

	mockRepo := &mockUsuarioRepository{}
	mockCargoRepo := &mockCargoRepository{}
	mockEmpresaRepo := &mockEmpresaRepository{}
	service := NewUsuarioService(mockRepo, mockCargoRepo, mockEmpresaRepo)
	usuarioParaCriar := &model.Usuario{
		Email:     "existente@email.com",
		EmpresaID: 1,
		CargoID:   2,
	}
	mockRepo.FindByEmailFunc = func(email string) (*model.Usuario, error) {
		return &model.Usuario{ID: 1, Email: "existente@email.com"}, nil
	}
	mockEmpresaRepo.FindByIDFunc = func(id uint) (*model.Empresa, error) {
		return &model.Empresa{ID: id, Nome: "Empresa Teste"}, nil
	}
	mockCargoRepo.FindByIDFunc = func(id uint, empresaID uint) (*model.Cargo, error) {
		return &model.Cargo{ID: id, EmpresaID: empresaID, Nome: "Testador"}, nil
	}
	err := service.CriarUsuario(usuarioParaCriar, "")
	if err == nil {
		t.Error("Esperado um erro de e-mail já cadastrado, mas nenhum erro foi retornado")
	}
	expectedErrorMsg := "e-mail já cadastrado"
	if err != nil && err.Error() != expectedErrorMsg {
		t.Errorf("Mensagem de erro incorreta. Esperado: '%s', Recebido: '%s'", expectedErrorMsg, err.Error())
	}
}

func TestCriarUsuario_ErroNaCriptografia(t *testing.T) {

	mockRepo := &mockUsuarioRepository{}
	mockCargoRepo := &mockCargoRepository{}
	mockEmpresaRepo := &mockEmpresaRepository{}
	service := NewUsuarioService(mockRepo, mockCargoRepo, mockEmpresaRepo)

	usuarioParaCriar := &model.Usuario{
		Nome:      "Usuário de Teste",
		Email:     "cripto@email.com",
		Senha:     "senha123",
		EmpresaID: 1,
		CargoID:   2,
	}
	mockRepo.FindByEmailFunc = func(email string) (*model.Usuario, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockEmpresaRepo.FindByIDFunc = func(id uint) (*model.Empresa, error) {
		return &model.Empresa{ID: id, Nome: "Empresa Teste"}, nil
	}
	mockCargoRepo.FindByIDFunc = func(id uint, empresaID uint) (*model.Cargo, error) {
		return &model.Cargo{ID: id, EmpresaID: empresaID, Nome: "Testador"}, nil
	}
	originalCriptografaSenha := criptografaSenha
	defer func() { criptografaSenha = originalCriptografaSenha }()
	expectedError := errors.New("erro simulado na criptografia")
	criptografaSenha = func(senha string) (string, error) {
		return "", expectedError
	}
	err := service.CriarUsuario(usuarioParaCriar, "")
	if err == nil {
		t.Fatal("Esperado um erro de criptografia, mas nenhum erro foi retornado")
	}
	if err.Error() != expectedError.Error() {
		t.Errorf("Mensagem de erro incorreta. Esperado: '%s', Recebido: '%s'", expectedError.Error(), err.Error())
	}
}

func TestGetAll_ComSucesso(t *testing.T) {
	// 1. Configurar o mock e o serviço

	mockRepo := &mockUsuarioRepository{}
	mockCargoRepo := &mockCargoRepository{}
	mockEmpresaRepo := &mockEmpresaRepository{}
	service := NewUsuarioService(mockRepo, mockCargoRepo, mockEmpresaRepo)
	empresaID := uint(1)
	usuariosEsperados := []model.Usuario{
		{ID: 1, Nome: "João", Email: "joao@email.com", EmpresaID: empresaID},
		{ID: 2, Nome: "Maria", Email: "maria@email.com", EmpresaID: empresaID},
	}
	mockRepo.GetAllFunc = func(empID uint) ([]model.Usuario, error) {
		if empID == empresaID {
			return usuariosEsperados, nil
		}
		return nil, errors.New("empresa não encontrada")
	}
	usuarios, err := service.GetAll(empresaID)
	if err != nil {
		t.Fatalf("Esperava não ter erro, mas recebeu: %v", err)
	}
	if len(usuarios) != len(usuariosEsperados) {
		t.Fatalf("Número de usuários incorreto. Esperava %d, mas recebeu %d", len(usuariosEsperados), len(usuarios))
	}
	for i, usuario := range usuarios {
		if usuario.ID != usuariosEsperados[i].ID {
			t.Errorf("ID incorreto. Esperava '%d', mas recebeu '%d'", usuariosEsperados[i].ID, usuario.ID)
		}
		if usuario.Nome != usuariosEsperados[i].Nome {
			t.Errorf("Nome incorreto. Esperava '%s', mas recebeu '%s'", usuariosEsperados[i].Nome, usuario.Nome)
		}
	}
}

func TestGetAll_ComErro(t *testing.T) {
	// 1. Configurar o mock e o serviço

	mockRepo := &mockUsuarioRepository{}
	mockCargoRepo := &mockCargoRepository{}
	mockEmpresaRepo := &mockEmpresaRepository{}
	service := NewUsuarioService(mockRepo, mockCargoRepo, mockEmpresaRepo)
	empresaID := uint(1)
	expectedError := errors.New("erro de banco de dados simulado")
	mockRepo.GetAllFunc = func(empID uint) ([]model.Usuario, error) {
		return nil, expectedError
	}
	_, err := service.GetAll(empresaID)
	if err == nil {
		t.Fatal("Esperava um erro, mas não recebeu nenhum")
	}
	if err.Error() != expectedError.Error() {
		t.Errorf("Mensagem de erro incorreta. Esperado: '%s', Recebido: '%s'", expectedError.Error(), err.Error())
	}
}

func TestUpdate_UsuarioNaoEncontrado(t *testing.T) {
	// 1. Configurar o mock e o serviço

	mockRepo := &mockUsuarioRepository{}
	mockCargoRepo := &mockCargoRepository{}
	mockEmpresaRepo := &mockEmpresaRepository{}
	service := NewUsuarioService(mockRepo, mockCargoRepo, mockEmpresaRepo)
	usuarioID := uint(999)
	empresaID := uint(1)
	dadosParaAtualizar := map[string]interface{}{"nome": "Usuário Atualizado"}
	mockRepo.FindByIDFunc = func(id uint, empID uint) (*model.Usuario, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockRepo.UpdateFunc = func(id uint, empID uint, dados map[string]interface{}) error {
		t.Errorf("Método Update do repositório foi chamado, mas não deveria")
		return nil
	}
	err := service.Update(usuarioID, empresaID, dadosParaAtualizar)
	if err == nil {
		t.Fatal("Esperava um erro de 'registro não encontrado', mas não recebeu nenhum")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Tipo de erro incorreto. Esperava 'gorm.ErrRecordNotFound', mas recebeu '%v'", err)
	}
}

func TestDelete_UsuarioNaoEncontrado(t *testing.T) {
	// 1. Configurar o mock e o serviço

	mockRepo := &mockUsuarioRepository{}
	mockCargoRepo := &mockCargoRepository{}
	mockEmpresaRepo := &mockEmpresaRepository{}
	service := NewUsuarioService(mockRepo, mockCargoRepo, mockEmpresaRepo)
	usuarioID := uint(999)
	empresaID := uint(1)
	mockRepo.FindByIDFunc = func(id uint, empID uint) (*model.Usuario, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockRepo.DeleteFunc = func(id uint, empID uint) error {
		t.Errorf("Método Delete do repositório foi chamado, mas não deveria")
		return nil
	}
	err := service.Delete(usuarioID, empresaID)
	if err == nil {
		t.Fatal("Esperava um erro de 'registro não encontrado', mas não recebeu nenhum")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Tipo de erro incorreto. Esperava 'gorm.ErrRecordNotFound', mas recebeu '%v'", err)
	}
}
