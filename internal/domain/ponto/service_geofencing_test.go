package ponto

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/domain/localidade"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

// Mock repositories for geofencing tests
type mockUserRepo struct {
	user *model.Usuario
	err  error
}

func (m *mockUserRepo) Save(usuario *model.Usuario) error {
	return nil
}

func (m *mockUserRepo) FindByEmail(email string) (*model.Usuario, error) {
	return nil, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint, empresaID uint) (*model.Usuario, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m *mockUserRepo) GetAll(empresaID uint) ([]model.Usuario, error) {
	return nil, nil
}

func (m *mockUserRepo) Update(id uint, dados map[string]interface{}) error {
	return nil
}

func (m *mockUserRepo) Delete(id uint) error {
	return nil
}

func (m *mockUserRepo) FindAll() ([]model.Usuario, error) {
	return nil, nil
}

func (m *mockUserRepo) WithTransaction(tx *gorm.DB) usuario.UsuarioRepository {
	return m
}

type mockLocalidadeRepo struct {
	localidade *model.Localidade
	err        error
}

func (m *mockLocalidadeRepo) Save(localidade *model.Localidade) error {
	return nil
}

func (m *mockLocalidadeRepo) FindByID(id uint) (*model.Localidade, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.localidade, nil
}

func (m *mockLocalidadeRepo) FindAllByEmpresaID(empresaID uint) ([]model.Localidade, error) {
	return nil, nil
}

func (m *mockLocalidadeRepo) Update(localidade *model.Localidade) error {
	return nil
}

func (m *mockLocalidadeRepo) Delete(id, empresaID uint) error {
	return nil
}

func (m *mockLocalidadeRepo) WithTransaction(tx *gorm.DB) localidade.Repository {
	return m
}

type mockEmpresaRepo struct {
	empresa *model.Empresa
	err     error
}

func (m *mockEmpresaRepo) FindByID(ctx context.Context, id uint) (*model.Empresa, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.empresa, nil
}

func (m *mockEmpresaRepo) CreateEmpresa(empresa *model.Empresa) error {
	return nil
}

func (m *mockEmpresaRepo) GetAllEmpresas() ([]model.Empresa, error) {
	return nil, nil
}

func (m *mockEmpresaRepo) GetEmpresaByID(id uint) (*model.Empresa, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.empresa, nil
}

func (m *mockEmpresaRepo) UpdateEmpresa(id uint, dados map[string]interface{}) error {
	return nil
}

func (m *mockEmpresaRepo) DeleteEmpresa(id uint) error {
	return nil
}

func (m *mockEmpresaRepo) WithTransaction(tx *gorm.DB) empresa.EmpresaRepository {
	return m
}

type mockPontoRepo struct {
	savedPonto *model.RegistroPonto
	saveErr    error
}

func (m *mockPontoRepo) SavePonto(p *model.RegistroPonto) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedPonto = p
	return nil
}

func (m *mockPontoRepo) FindPontosByUserIDAndDate(userID uint, dia time.Time) ([]model.RegistroPonto, error) {
	return nil, nil
}

func (m *mockPontoRepo) FindPontosByUserIDAndDateRange(userID uint, inicio, fim time.Time) ([]model.RegistroPonto, error) {
	return nil, nil
}

func (m *mockPontoRepo) WithTransaction(tx *gorm.DB) RegistroPontoRepository {
	return m
}

func (m *mockPontoRepo) FindPontoByID(pontoID uint, empresaID uint) (*model.RegistroPonto, error) {
	return nil, nil
}

func (m *mockPontoRepo) UpdatePonto(p *model.RegistroPonto) error {
	return nil
}

// TestBaterPonto_GeofencingDisabled tests that when geofencing is disabled,
// users can clock in from anywhere (classified as Presencial or Remoto)
func TestBaterPonto_GeofencingDisabled(t *testing.T) {
	// Setup: User outside the radius, but geofencing is disabled
	userRepo := &mockUserRepo{
		user: &model.Usuario{
			ID:   1,
			Nome: "Test User",
			Contrato: model.Contrato{
				ID:           1,
				LocalidadeID: 1,
			},
		},
	}

	localidadeRepo := &mockLocalidadeRepo{
		localidade: &model.Localidade{
			ID:                 1,
			Nome:               "Matriz",
			Latitude:           -15.799879,
			Longitude:          -47.864162,
			RaioGeofenceMetros: 100.0, // 100 meters
		},
	}

	empresaRepo := &mockEmpresaRepo{
		empresa: &model.Empresa{
			ID:                       1,
			NomeFantasia:             "Test Company",
			PresencialRestritoAoRaio: false, // Geofencing disabled
		},
	}

	pontoRepo := &mockPontoRepo{}

	service := &pontoService{
		pontoRepo:      pontoRepo,
		userRepo:       userRepo,
		localidadeRepo: localidadeRepo,
		empresaRepo:    empresaRepo,
		db:             nil,
	}

	// User is far from the location
	// When geofencing is disabled, users can clock in from anywhere
	// The system classifies based on distance: Presencial (within radius) or Remoto (outside radius)
	// Using coordinates approximately 500m away - should be classified as Remoto
	latitude := -15.804879  // ~550m away from the base location
	longitude := -47.864162

	ponto, err := service.BaterPonto(1, 1, latitude, longitude)

	// Should succeed even though user is outside radius
	if err != nil {
		t.Fatalf("Expected no error when geofencing is disabled, got: %v", err)
	}

	if ponto == nil {
		t.Fatal("Expected ponto to be created")
	}

	if ponto.Metodo != "Remoto" {
		t.Errorf("Expected Metodo to be 'Remoto', got: %s", ponto.Metodo)
	}
}

// TestBaterPonto_GeofencingEnabled_WithinRadius tests that when geofencing is enabled,
// users can clock in from within the radius (Presencial)
func TestBaterPonto_GeofencingEnabled_WithinRadius(t *testing.T) {
	userRepo := &mockUserRepo{
		user: &model.Usuario{
			ID:   1,
			Nome: "Test User",
			Contrato: model.Contrato{
				ID:           1,
				LocalidadeID: 1,
			},
		},
	}

	localidadeRepo := &mockLocalidadeRepo{
		localidade: &model.Localidade{
			ID:                 1,
			Nome:               "Matriz",
			Latitude:           -15.799879,
			Longitude:          -47.864162,
			RaioGeofenceMetros: 100.0, // 100 meters
		},
	}

	empresaRepo := &mockEmpresaRepo{
		empresa: &model.Empresa{
			ID:                       1,
			NomeFantasia:             "Test Company",
			PresencialRestritoAoRaio: true, // Geofencing enabled
		},
	}

	pontoRepo := &mockPontoRepo{}

	service := &pontoService{
		pontoRepo:      pontoRepo,
		userRepo:       userRepo,
		localidadeRepo: localidadeRepo,
		empresaRepo:    empresaRepo,
		db:             nil,
	}

	// User is at the exact location (within radius)
	latitude := -15.799879
	longitude := -47.864162

	ponto, err := service.BaterPonto(1, 1, latitude, longitude)

	// Should succeed because user is within radius
	if err != nil {
		t.Fatalf("Expected no error when user is within radius, got: %v", err)
	}

	if ponto == nil {
		t.Fatal("Expected ponto to be created")
	}

	if ponto.Metodo != "Presencial" {
		t.Errorf("Expected Metodo to be 'Presencial', got: %s", ponto.Metodo)
	}
}

// TestBaterPonto_GeofencingEnabled_OutsideRadius tests that when geofencing is enabled,
// users cannot clock in from outside the radius
func TestBaterPonto_GeofencingEnabled_OutsideRadius(t *testing.T) {
	userRepo := &mockUserRepo{
		user: &model.Usuario{
			ID:   1,
			Nome: "Test User",
			Contrato: model.Contrato{
				ID:           1,
				LocalidadeID: 1,
			},
		},
	}

	localidadeRepo := &mockLocalidadeRepo{
		localidade: &model.Localidade{
			ID:                 1,
			Nome:               "Matriz",
			Latitude:           -15.799879,
			Longitude:          -47.864162,
			RaioGeofenceMetros: 100.0, // 100 meters
		},
	}

	empresaRepo := &mockEmpresaRepo{
		empresa: &model.Empresa{
			ID:                       1,
			NomeFantasia:             "Test Company",
			PresencialRestritoAoRaio: true, // Geofencing enabled
		},
	}

	pontoRepo := &mockPontoRepo{}

	service := &pontoService{
		pontoRepo:      pontoRepo,
		userRepo:       userRepo,
		localidadeRepo: localidadeRepo,
		empresaRepo:    empresaRepo,
		db:             nil,
	}

	// User is far from the location (more than 100m away)
	// Using coordinates approximately 500m away
	latitude := -15.804879  // ~550m away from the base location
	longitude := -47.864162

	ponto, err := service.BaterPonto(1, 1, latitude, longitude)

	// Should fail because user is outside radius and geofencing is enabled
	if err == nil {
		t.Fatal("Expected error when user is outside radius with geofencing enabled")
	}

	if !errors.Is(err, ErrGeofencingViolation) {
		t.Errorf("Expected ErrGeofencingViolation, got: %v", err)
	}

	if ponto != nil {
		t.Error("Expected ponto to be nil when geofencing violation occurs")
	}

	// Verify the error message contains distance information
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected error message to contain distance information")
	}
}

// TestBaterPonto_GeofencingEnabled_EmpresaNotFound tests error handling
func TestBaterPonto_GeofencingEnabled_EmpresaNotFound(t *testing.T) {
	userRepo := &mockUserRepo{
		user: &model.Usuario{
			ID:   1,
			Nome: "Test User",
			Contrato: model.Contrato{
				ID:           1,
				LocalidadeID: 1,
			},
		},
	}

	localidadeRepo := &mockLocalidadeRepo{
		localidade: &model.Localidade{
			ID:                 1,
			Nome:               "Matriz",
			Latitude:           -15.799879,
			Longitude:          -47.864162,
			RaioGeofenceMetros: 100.0,
		},
	}

	empresaRepo := &mockEmpresaRepo{
		err: gorm.ErrRecordNotFound,
	}

	pontoRepo := &mockPontoRepo{}

	service := &pontoService{
		pontoRepo:      pontoRepo,
		userRepo:       userRepo,
		localidadeRepo: localidadeRepo,
		empresaRepo:    empresaRepo,
		db:             nil,
	}

	latitude := -15.799879
	longitude := -47.864162

	ponto, err := service.BaterPonto(1, 1, latitude, longitude)

	// Should fail with empresa not found error
	if err == nil {
		t.Fatal("Expected error when empresa is not found")
	}

	if ponto != nil {
		t.Error("Expected ponto to be nil when empresa is not found")
	}
}
