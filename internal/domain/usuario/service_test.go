package usuario

import (
	"errors"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
	"testing"
)

type mockUsuarioRepository struct {
	SaveFunc        func(usuario *model.Usuario) error
	FindByEmailFunc func(email string) (*model.Usuario, error)
	FindByIDFunc    func(id uint) (*model.Usuario, error)
	GetAllFunc      func() ([]model.Usuario, error)
	UpdateFunc      func(id uint, dados map[string]interface{}) error
	DeleteFunc      func(id uint) error
	CriarFunc       func(usuario *model.Usuario) error
}

func (m *mockUsuarioRepository) Save(usuario *model.Usuario) error {
	return m.SaveFunc(usuario)
}

func (m *mockUsuarioRepository) FindByEmail(email string) (*model.Usuario, error) {
	return m.FindByEmailFunc(email)
}

func (m *mockUsuarioRepository) FindByID(id uint) (*model.Usuario, error) {
	return m.FindByIDFunc(id)
}

func (m *mockUsuarioRepository) GetAll() ([]model.Usuario, error) {
	return m.GetAllFunc()
}

func (m *mockUsuarioRepository) Update(id uint, dados map[string]interface{}) error {
	return m.UpdateFunc(id, dados)
}

func (m *mockUsuarioRepository) Delete(id uint) error {
	return m.DeleteFunc(id)
}

func TestCriarUsuario_ComSucesso(t *testing.T) {
	mockRepo := &mockUsuarioRepository{}

	service := NewUsuarioService(mockRepo)

	usuarioParaCriar := &model.Usuario{
		Nome:  "Usuário de Teste",
		Email: "sucesso@email.com",
		Senha: "senha123",
		Cargo: "Testador",
	}

	mockRepo.FindByEmailFunc = func(email string) (*model.Usuario, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockRepo.SaveFunc = func(usuario *model.Usuario) error {
		return nil
	}

	err := service.CriarUsuario(usuarioParaCriar)
	if err != nil {
		t.Errorf("Erro inesperado ao criar usuário: %v", err)
	}
}

func TestFindByIDComSucesso(t *testing.T) {
	mockRepo := &mockUsuarioRepository{}
	service := NewUsuarioService(mockRepo)

	usuarioID := uint(1)
	usuarioEsperado := &model.Usuario{
		ID:    usuarioID,
		Nome:  "Usuário de Teste",
		Email: "sucesso@email.com",
		Senha: "senha123",
		Cargo: "Testador",
	}

	mockRepo.FindByIDFunc = func(id uint) (*model.Usuario, error) {
		if id == usuarioID {
			return usuarioEsperado, nil
		}
		return nil, gorm.ErrRecordNotFound
	}

	usuario, err := service.FindByID(usuarioID)

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
	service := NewUsuarioService(mockRepo)
	usuarioParaCriar := &model.Usuario{
		Email: "existente@email.com",
	}

	mockRepo.FindByEmailFunc = func(email string) (*model.Usuario, error) {
		return &model.Usuario{ID: 1, Email: "existente@email.com"}, nil
	}

	err := service.CriarUsuario(usuarioParaCriar)

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
	service := NewUsuarioService(mockRepo)

	usuarioParaCriar := &model.Usuario{
		Nome:  "Usuário de Teste",
		Email: "cripto@email.com",
		Senha: "senha123",
		Cargo: "Testador",
	}

	mockRepo.FindByEmailFunc = func(email string) (*model.Usuario, error) {
		return nil, gorm.ErrRecordNotFound
	}

	originalCriptografaSenha := criptografaSenha
	defer func() { criptografaSenha = originalCriptografaSenha }()

	expectedError := errors.New("erro simulado na criptografia")
	criptografaSenha = func(senha string) (string, error) {
		return "", expectedError
	}

	err := service.CriarUsuario(usuarioParaCriar)

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
	service := NewUsuarioService(mockRepo)

	// Lista de usuários simulada
	usuariosEsperados := []model.Usuario{
		{ID: 1, Nome: "João", Email: "joao@email.com"},
		{ID: 2, Nome: "Maria", Email: "maria@email.com"},
	}

	// 2. Definir o comportamento esperado do mock
	// Esperamos que GetAllFunc retorne a lista de usuários e nenhum erro.
	mockRepo.GetAllFunc = func() ([]model.Usuario, error) {
		return usuariosEsperados, nil
	}

	// 3. Executar o método do serviço
	usuarios, err := service.GetAll()

	// 4. Fazer as verificações (assertions)
	if err != nil {
		t.Fatalf("Esperava não ter erro, mas recebeu: %v", err)
	}

	if len(usuarios) != len(usuariosEsperados) {
		t.Fatalf("Número de usuários incorreto. Esperava %d, mas recebeu %d", len(usuariosEsperados), len(usuarios))
	}

	// 5. Verificar se os dados retornados são os esperados
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
	service := NewUsuarioService(mockRepo)

	// 2. Definir o comportamento esperado do mock
	// Esperamos que GetAllFunc retorne um erro simulado.
	expectedError := errors.New("erro de banco de dados simulado")
	mockRepo.GetAllFunc = func() ([]model.Usuario, error) {
		return nil, expectedError
	}

	// 3. Executar o método do serviço
	_, err := service.GetAll()

	// 4. Fazer as verificações (assertions)
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
	service := NewUsuarioService(mockRepo)

	usuarioID := uint(999) // Um ID que sabemos que não existe
	dadosParaAtualizar := map[string]interface{}{"nome": "Usuário Atualizado"}

	// 2. Definir o comportamento esperado do mock
	// Esperamos que FindByIDFunc retorne gorm.ErrRecordNotFound,
	// simulando que o usuário não foi encontrado no banco de dados.
	mockRepo.FindByIDFunc = func(id uint) (*model.Usuario, error) {
		return nil, gorm.ErrRecordNotFound
	}

	// 3. Executar o método do serviço
	err := service.Update(usuarioID, dadosParaAtualizar)

	// 4. Fazer as verificações (assertions)
	if err == nil {
		t.Fatal("Esperava um erro de 'registro não encontrado', mas não recebeu nenhum")
	}

	// Verificamos se o erro é exatamente o que esperávamos.
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Tipo de erro incorreto. Esperava 'gorm.ErrRecordNotFound', mas recebeu '%v'", err)
	}

	// O método Update do repositório não deve ser chamado
	// se o FindByID falhar primeiro.
	if mockRepo.UpdateFunc != nil {
		t.Errorf("Método Update do repositório foi chamado, mas não deveria")
	}
}

func TestDelete_UsuarioNaoEncontrado(t *testing.T) {
	// 1. Configurar o mock e o serviço
	mockRepo := &mockUsuarioRepository{}
	service := NewUsuarioService(mockRepo)

	usuarioID := uint(999) // Um ID que sabemos que não existe

	// 2. Definir o comportamento esperado do mock
	// Esperamos que FindByIDFunc retorne gorm.ErrRecordNotFound,
	// simulando que o usuário não foi encontrado.
	mockRepo.FindByIDFunc = func(id uint) (*model.Usuario, error) {
		return nil, gorm.ErrRecordNotFound
	}

	// 3. Executar o método do serviço
	err := service.Delete(usuarioID)

	// 4. Fazer as verificações (assertions)
	if err == nil {
		t.Fatal("Esperava um erro de 'registro não encontrado', mas não recebeu nenhum")
	}

	// Verificamos se o erro é o esperado.
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Tipo de erro incorreto. Esperava 'gorm.ErrRecordNotFound', mas recebeu '%v'", err)
	}

	// O método Delete do repositório não deve ser chamado
	// se o FindByID falhar primeiro.
	if mockRepo.DeleteFunc != nil {
		t.Errorf("Método Delete do repositório foi chamado, mas não deveria")
	}
}
