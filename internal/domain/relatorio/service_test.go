package relatorio

import (
	"context"
	"testing"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/bancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/justificativa"
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
type mockUserRepo struct {
	carga int
	users []model.Usuario
}

func (r *mockUserRepo) FindByID(ctx context.Context, id uint, empresaID uint) (*model.Usuario, error) {
	// Se tem lista de usuários, retorna da lista
	for _, u := range r.users {
		if u.ID == id {
			return &u, nil
		}
	}
	// Fallback para comportamento antigo
	return &model.Usuario{
		ID:    id,
		Nome:  "Usuário Teste",
		CPF:   "12345678900",
		Email: "teste@example.com",
		Contrato: model.Contrato{
			ID:        1,
			EmpresaID: empresaID,
			Cargo: model.Cargo{
				ID:                        1,
				Nome:                      "Desenvolvedor",
				CargaHorariaDiariaMinutos: uint(r.carga),
			},
		},
	}, nil
}

func (r *mockUserRepo) GetAllActive(empresaID uint) ([]model.Usuario, error) {
	return r.users, nil
}

type mockJustificativaRepo struct {
	justificativas []model.Justificativa
}

func (m *mockJustificativaRepo) Create(j *model.Justificativa) error { return nil }
func (m *mockJustificativaRepo) FindByID(id uint, empresaID uint) (*model.Justificativa, error) {
	return nil, nil
}
func (m *mockJustificativaRepo) FindByStatus(empresaID uint, status string) ([]model.Justificativa, error) {
	return nil, nil
}
func (m *mockJustificativaRepo) FindByUsuarioIDAndPeriodo(usuarioID uint, empresaID uint, inicio, fim time.Time) ([]model.Justificativa, error) {
	var result []model.Justificativa
	for _, j := range m.justificativas {
		if j.UsuarioID == usuarioID && !j.DataOcorrencia.Before(inicio) && !j.DataOcorrencia.After(fim) {
			result = append(result, j)
		}
	}
	return result, nil
}
func (m *mockJustificativaRepo) Update(j *model.Justificativa) error { return nil }
func (m *mockJustificativaRepo) WithTransaction(tx *gorm.DB) justificativa.Repository {
	return m
}

type mockLogRepo struct{}

func (m *mockLogRepo) Create(l *model.LogBancoHoras) error                  { return nil }
func (m *mockLogRepo) WithTransaction(tx *gorm.DB) logbancohoras.Repository { return m }
func (m *mockLogRepo) GetAllByUsuarioAndEmpresa(usuarioID uint, empresaID uint) ([]model.LogBancoHoras, error) {
	return nil, nil
}

type mockBancoHorasService struct{ carga int }

func (m *mockBancoHorasService) CalcularSaldoParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (int, error) {
	// Recalcula a partir das marcações providas pelo mockPontoRepo? Simplesmente usa a carga para um resultado determinístico nos testes.
	// Para manter compatibilidade com os testes existentes que avaliam apenas Totais.TrabalhadoMinutos,
	// podemos retornar 0 aqui e o serviço de relatorio ainda computa totalTrabalhado para Totais.
	return 0, nil
}
func (m *mockBancoHorasService) FecharDiaParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (*model.Contrato, error) {
	return &model.Contrato{}, nil
}
func (m *mockBancoHorasService) GetDashboardForUsuario(usuarioID uint, empresaID uint) (*bancohoras.DashboardResponse, error) {
	return &bancohoras.DashboardResponse{}, nil
}
func (m *mockBancoHorasService) GetSaldoAtualUsuario(usuarioID uint, empresaID uint) (int, error) {
	return 0, nil
}
func (m *mockBancoHorasService) InvalidarCacheDia(usuarioID uint, empresaID uint, dia time.Time) {
	// Mock não faz nada
}
func (m *mockBancoHorasService) InvalidarCacheUsuario(usuarioID uint, empresaID uint) {
	// Mock não faz nada
}

func buildService(regs []model.RegistroPonto, carga int) Service {
	return NewService(&mockPontoRepo{regs: regs}, &mockUserRepo{carga: carga}, &mockLogRepo{}, &mockBancoHorasService{carga: carga})
}

func buildServiceWithJustificativa(regs []model.RegistroPonto, users []model.Usuario, justificativas []model.Justificativa, carga int) Service {
	return NewServiceWithJustificativa(
		&mockPontoRepo{regs: regs},
		&mockUserRepo{carga: carga, users: users},
		&mockLogRepo{},
		&mockJustificativaRepo{justificativas: justificativas},
		&mockBancoHorasService{carga: carga},
	)
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

// --- Testes para GerarRelatorioGeral ---

func TestRelatorioGeralUsuarioEspecifico(t *testing.T) {
	t.Parallel()
	base := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)

	// Criar usuário de teste
	usuario := model.Usuario{
		ID:    10,
		Nome:  "João Silva",
		CPF:   "12345678900",
		Email: "joao@example.com",
		Contrato: model.Contrato{
			ID:                     1,
			EmpresaID:              1,
			SaldoBancoHorasMinutos: 0,
			Cargo: model.Cargo{
				ID:                        1,
				Nome:                      "Desenvolvedor",
				CargaHorariaDiariaMinutos: 480, // 8 horas
			},
		},
	}

	// Registros de ponto: dia 1 com 8h, dia 2 com 10h (2h extras)
	regs := []model.RegistroPonto{
		{ID: 1, UsuarioID: 10, Timestamp: base.Add(8 * time.Hour)},                      // Entrada dia 1
		{ID: 2, UsuarioID: 10, Timestamp: base.Add(17 * time.Hour)},                     // Saída dia 1 (9h depois)
		{ID: 3, UsuarioID: 10, Timestamp: base.Add(24 * time.Hour).Add(8 * time.Hour)},  // Entrada dia 2
		{ID: 4, UsuarioID: 10, Timestamp: base.Add(24 * time.Hour).Add(18 * time.Hour)}, // Saída dia 2 (10h depois)
	}

	usuarios := []model.Usuario{usuario}
	svc := buildServiceWithJustificativa(regs, usuarios, nil, 480)

	uid := uint(10)
	userIDLogado := uint(10) // O mesmo usuário acessando seus próprios dados
	relatorio, err := svc.GerarRelatorioGeral(base, base.Add(24*time.Hour), &uid, 1, userIDLogado)

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(relatorio.Usuarios) != 1 {
		t.Fatalf("esperava 1 usuário, obtido %d", len(relatorio.Usuarios))
	}

	usuarioRel := relatorio.Usuarios[0]
	if usuarioRel.Usuario.Nome != "João Silva" {
		t.Errorf("nome incorreto: %s", usuarioRel.Usuario.Nome)
	}

	if len(usuarioRel.Dias) != 2 {
		t.Fatalf("esperava 2 dias, obtido %d", len(usuarioRel.Dias))
	}

	// Verificar total trabalhado (9h + 10h = 19h = 1140 min)
	totalEsperado := 19 * 60
	if usuarioRel.Resumo.TotalHorasTrabalhadasMinutos != totalEsperado {
		t.Errorf("total trabalhado incorreto: esperado %d, obtido %d",
			totalEsperado, usuarioRel.Resumo.TotalHorasTrabalhadasMinutos)
	}

	// Verificar horas extras (1h dia 1 + 2h dia 2 = 3h = 180 min)
	extrasEsperadas := 3 * 60
	if usuarioRel.Resumo.TotalHorasExtrasMinutos != extrasEsperadas {
		t.Errorf("horas extras incorretas: esperado %d, obtido %d",
			extrasEsperadas, usuarioRel.Resumo.TotalHorasExtrasMinutos)
	}
}

func TestRelatorioGeralComJustificativa(t *testing.T) {
	t.Parallel()
	base := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)

	usuario := model.Usuario{
		ID:    10,
		Nome:  "Maria Santos",
		CPF:   "98765432100",
		Email: "maria@example.com",
		Contrato: model.Contrato{
			ID:                     1,
			EmpresaID:              1,
			SaldoBancoHorasMinutos: 0,
			Cargo: model.Cargo{
				ID:                        1,
				Nome:                      "Analista",
				CargaHorariaDiariaMinutos: 480, // 8 horas
			},
		},
	}

	// Dia 1: apenas 4 horas trabalhadas (faltou 4h)
	regs := []model.RegistroPonto{
		{ID: 1, UsuarioID: 10, Timestamp: base.Add(8 * time.Hour)},
		{ID: 2, UsuarioID: 10, Timestamp: base.Add(12 * time.Hour)},
	}

	// Justificativa aprovada para o dia 1
	justificativas := []model.Justificativa{
		{
			ID:             1,
			UsuarioID:      10,
			EmpresaID:      1,
			DataOcorrencia: base,
			Tipo:           "ATESTADO_MEDICO",
			Descricao:      "Consulta médica",
			Status:         "APROVADA",
		},
	}

	usuarios := []model.Usuario{usuario}
	svc := buildServiceWithJustificativa(regs, usuarios, justificativas, 480)

	uid := uint(10)
	userIDLogado := uint(10) // O mesmo usuário acessando seus próprios dados
	relatorio, err := svc.GerarRelatorioGeral(base, base, &uid, 1, userIDLogado)

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	usuarioRel := relatorio.Usuarios[0]

	// Com justificativa aprovada, não deve contar como horas faltantes
	if usuarioRel.Resumo.TotalHorasFaltantesMinutos != 0 {
		t.Errorf("não deveria ter horas faltantes com justificativa aprovada, obtido %d",
			usuarioRel.Resumo.TotalHorasFaltantesMinutos)
	}

	// Verificar que a justificativa está presente no dia
	if len(usuarioRel.Dias) != 1 {
		t.Fatalf("esperava 1 dia")
	}
	dia := usuarioRel.Dias[0]
	if dia.Justificativa == nil {
		t.Error("justificativa deveria estar presente")
	} else if dia.Justificativa.Status != "APROVADA" {
		t.Errorf("status da justificativa incorreto: %s", dia.Justificativa.Status)
	}
}

func TestRelatorioGeralSemJustificativa(t *testing.T) {
	t.Parallel()
	base := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)

	usuario := model.Usuario{
		ID:    10,
		Nome:  "Pedro Costa",
		CPF:   "11122233344",
		Email: "pedro@example.com",
		Contrato: model.Contrato{
			ID:                     1,
			EmpresaID:              1,
			SaldoBancoHorasMinutos: 0,
			Cargo: model.Cargo{
				ID:                        1,
				Nome:                      "Designer",
				CargaHorariaDiariaMinutos: 480, // 8 horas
			},
		},
	}

	// Dia 1: apenas 6 horas trabalhadas (faltou 2h, sem justificativa)
	regs := []model.RegistroPonto{
		{ID: 1, UsuarioID: 10, Timestamp: base.Add(8 * time.Hour)},
		{ID: 2, UsuarioID: 10, Timestamp: base.Add(14 * time.Hour)},
	}

	usuarios := []model.Usuario{usuario}
	svc := buildServiceWithJustificativa(regs, usuarios, nil, 480)

	uid := uint(10)
	userIDLogado := uint(10) // O mesmo usuário acessando seus próprios dados
	relatorio, err := svc.GerarRelatorioGeral(base, base, &uid, 1, userIDLogado)

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	usuarioRel := relatorio.Usuarios[0]

	// Sem justificativa, deve contar 2h como faltantes
	faltantesEsperadas := 2 * 60
	if usuarioRel.Resumo.TotalHorasFaltantesMinutos != faltantesEsperadas {
		t.Errorf("horas faltantes incorretas: esperado %d, obtido %d",
			faltantesEsperadas, usuarioRel.Resumo.TotalHorasFaltantesMinutos)
	}
}

func TestRelatorioGeralMultiplosUsuarios(t *testing.T) {
	t.Parallel()
	base := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)

	usuarios := []model.Usuario{
		{
			ID:    10,
			Nome:  "Usuário 1",
			CPF:   "11111111111",
			Email: "user1@example.com",
			Contrato: model.Contrato{
				ID:                     1,
				EmpresaID:              1,
				SaldoBancoHorasMinutos: 0,
				Cargo: model.Cargo{
					ID:                        1,
					Nome:                      "Dev",
					CargaHorariaDiariaMinutos: 480,
				},
			},
		},
		{
			ID:    11,
			Nome:  "Usuário 2",
			CPF:   "22222222222",
			Email: "user2@example.com",
			Contrato: model.Contrato{
				ID:                     2,
				EmpresaID:              1,
				SaldoBancoHorasMinutos: 0,
				Cargo: model.Cargo{
					ID:                        1,
					Nome:                      "Dev",
					CargaHorariaDiariaMinutos: 480,
				},
			},
		},
	}

	regs := []model.RegistroPonto{
		// Usuário 1
		{ID: 1, UsuarioID: 10, Timestamp: base.Add(8 * time.Hour)},
		{ID: 2, UsuarioID: 10, Timestamp: base.Add(16 * time.Hour)},
		// Usuário 2
		{ID: 3, UsuarioID: 11, Timestamp: base.Add(9 * time.Hour)},
		{ID: 4, UsuarioID: 11, Timestamp: base.Add(18 * time.Hour)},
	}

	svc := buildServiceWithJustificativa(regs, usuarios, nil, 480)

	// Usar um usuário com nível hierárquico alto (SuperAdmin) para acessar todos
	userIDLogado := uint(999) // ID fictício de um superadmin que não valida hierarquia no teste
	relatorio, err := svc.GerarRelatorioGeral(base, base, nil, 1, userIDLogado)

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(relatorio.Usuarios) != 2 {
		t.Fatalf("esperava 2 usuários, obtido %d", len(relatorio.Usuarios))
	}
}
