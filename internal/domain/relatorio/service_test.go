package relatorio

import (
	"testing"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/logbancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

// Implement full interfaces expected by relatorio service constructor.
type mockPontoRepo struct{ regs []model.RegistroPonto }

func (m *mockPontoRepo) SavePonto(p *model.RegistroPonto) error { return nil }
func (m *mockPontoRepo) FindPontosByUserIDAndDate(userID uint, dia time.Time) ([]model.RegistroPonto, error) {
	return nil, nil
}
func (m *mockPontoRepo) FindPontosByUserIDAndDateRange(userID uint, inicio, fim time.Time) ([]model.RegistroPonto, error) {
	return m.regs, nil
}
func (m *mockPontoRepo) WithTransaction(tx *gorm.DB) ponto.RegistroPontoRepository { return m }
func (m *mockPontoRepo) FindPontoByID(pontoID uint, empresaID uint) (*model.RegistroPonto, error) {
	return &model.RegistroPonto{ID: pontoID}, nil
}
func (m *mockPontoRepo) UpdatePonto(p *model.RegistroPonto) error { return nil }

// Minimal usuario reader already defined in service.go as usuarioReader; implement it.
type mockUserRepo struct{ carga int }

func (r *mockUserRepo) FindByID(id uint, empresaID uint) (*model.Usuario, error) {
	return &model.Usuario{ID: id, Contrato: model.Contrato{ID: 1, EmpresaID: empresaID, Cargo: model.Cargo{ID: 1, CargaHorariaDiariaMinutos: uint(r.carga)}}}, nil
}

type mockLogRepo struct{}

func (m *mockLogRepo) Create(l *model.LogBancoHoras) error                  { return nil }
func (m *mockLogRepo) WithTransaction(tx *gorm.DB) logbancohoras.Repository { return m }

func buildService(regs []model.RegistroPonto, carga int) Service {
	return NewService(&mockPontoRepo{regs: regs}, &mockUserRepo{carga: carga}, &mockLogRepo{})
}

func TestEspelhoSemRegistros(t *testing.T) {
	t.Parallel()
	svc := buildService(nil, 480)
	inicio := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	fim := time.Date(2025, 10, 3, 0, 0, 0, 0, time.UTC)
	espelho, err := svc.GerarEspelhoPonto(10, 55, inicio, fim)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(espelho.Dias) != 3 {
		t.Fatalf("esperava 3 dias, obtido %d", len(espelho.Dias))
	}
	if espelho.Totais.TrabalhadoMinutos != 0 {
		t.Errorf("total trabalhado deve ser 0")
	}
}

func TestEspelhoMarcacaoImpar(t *testing.T) {
	t.Parallel()
	dia := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	regs := []model.RegistroPonto{
		{ID: 1, UsuarioID: 10, Timestamp: dia.Add(8 * time.Hour)},
		{ID: 2, UsuarioID: 10, Timestamp: dia.Add(9 * time.Hour)},
		{ID: 3, UsuarioID: 10, Timestamp: dia.Add(18 * time.Hour)},
	}
	svc := buildService(regs, 480)
	espelho, err := svc.GerarEspelhoPonto(10, 1, dia, dia)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(espelho.DiasInconsistentes) != 1 {
		t.Errorf("esperava 1 inconsistente, obtido %v", espelho.DiasInconsistentes)
	}
}

func TestEspelhoMultiDiaSaldo(t *testing.T) {
	t.Parallel()
	base := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	regs := []model.RegistroPonto{
		{ID: 1, UsuarioID: 10, Timestamp: base.Add(8 * time.Hour)},
		{ID: 2, UsuarioID: 10, Timestamp: base.Add(12 * time.Hour)},
		{ID: 3, UsuarioID: 10, Timestamp: base.Add(13 * time.Hour)},
		{ID: 4, UsuarioID: 10, Timestamp: base.Add(17 * time.Hour)},
		{ID: 5, UsuarioID: 10, Timestamp: base.Add(24 * time.Hour).Add(8 * time.Hour)},
		{ID: 6, UsuarioID: 10, Timestamp: base.Add(24 * time.Hour).Add(11 * time.Hour)},
		{ID: 7, UsuarioID: 10, Timestamp: base.Add(24 * time.Hour).Add(12 * time.Hour)},
		{ID: 8, UsuarioID: 10, Timestamp: base.Add(24 * time.Hour).Add(14 * time.Hour)},
	}
	svc := buildService(regs, 480)
	espelho, err := svc.GerarEspelhoPonto(10, 1, base, base.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	esperado := (8*60 + 5*60) // 8h no primeiro dia, 5h no segundo (pares 8-11, 12-14)
	if espelho.Totais.TrabalhadoMinutos != esperado {
		t.Errorf("total trabalhado incorreto: %d != %d", espelho.Totais.TrabalhadoMinutos, esperado)
	}
}

func TestEspelhoIntervaloInvertido(t *testing.T) {
	t.Parallel()
	inicio := time.Date(2025, 10, 5, 0, 0, 0, 0, time.UTC)
	fim := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	svc := buildService(nil, 480)
	espelho, err := svc.GerarEspelhoPonto(99, 1, inicio, fim)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(espelho.Dias) != 5 {
		t.Errorf("esperava 5 dias normalizados, obtido %d", len(espelho.Dias))
	}
}
