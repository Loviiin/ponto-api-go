package ponto

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

// mock implementations

type mockRepo struct {
	registros  []model.RegistroPonto
	lastInicio time.Time
	lastFim    time.Time
}

func (m *mockRepo) SavePonto(ctx context.Context, p *model.RegistroPonto) error { return nil }
func (m *mockRepo) FindPontosByUserIDAndDate(ctx context.Context, userID uint, dia time.Time) ([]model.RegistroPonto, error) {
	return nil, nil
}
func (m *mockRepo) FindPontosByUserIDAndDateRange(ctx context.Context, userID uint, inicio, fim time.Time) ([]model.RegistroPonto, error) {
	m.lastInicio = inicio
	m.lastFim = fim
	return m.registros, nil
}
func (m *mockRepo) WithTransaction(tx *gorm.DB) RegistroPontoRepository { return m }
func (m *mockRepo) FindPontoByID(ctx context.Context, pontoID uint, empresaID uint) (*model.RegistroPonto, error) {
	return &model.RegistroPonto{ID: pontoID}, nil
}
func (m *mockRepo) UpdatePonto(ctx context.Context, p *model.RegistroPonto) error { return nil }

// (unused repository interfaces omitted in tests to keep focus on report generation)

func TestGerarRelatorioCSV(t *testing.T) {
	inicio := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	fim := time.Date(2025, 10, 2, 23, 59, 59, 0, time.UTC)
	registros := []model.RegistroPonto{
		{ID: 1, UsuarioID: 10, Timestamp: inicio.Add(8 * time.Hour), Latitude: -10.1, Longitude: -50.2, Metodo: "Presencial"},
		{ID: 2, UsuarioID: 10, Timestamp: inicio.Add(17 * time.Hour), Latitude: -10.2, Longitude: -50.3, Metodo: "Remoto"},
	}
	ps := &pontoService{pontoRepo: &mockRepo{registros: registros}}
	bytes_, contentType, filename, err := ps.GerarRelatorio(context.Background(), 10, 1, inicio, fim, "csv")
	if err != nil {
		t.Fatalf("erro não esperado: %v", err)
	}
	if contentType != "text/csv" {
		t.Errorf("contentType esperado text/csv, obtido %s", contentType)
	}
	if !strings.HasSuffix(filename, ".csv") {
		t.Errorf("filename deve terminar com .csv: %s", filename)
	}
	data := string(bytes_)
	if !strings.Contains(data, "id,timestamp") {
		t.Errorf("csv deve conter cabeçalho; obtido: %s", data)
	}
	if !strings.Contains(data, "Presencial") || !strings.Contains(data, "Remoto") {
		t.Errorf("csv deve conter métodos")
	}
}

func TestGerarRelatorioPDF(t *testing.T) {
	inicio := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	fim := time.Date(2025, 10, 2, 23, 59, 59, 0, time.UTC)
	registros := []model.RegistroPonto{{ID: 1, UsuarioID: 10, Timestamp: inicio.Add(8 * time.Hour), Latitude: 1.2345, Longitude: 2.3456, Metodo: "Presencial"}}
	ps := &pontoService{pontoRepo: &mockRepo{registros: registros}}
	bytes_, contentType, filename, err := ps.GerarRelatorio(context.Background(), 10, 1, inicio, fim, "pdf")
	if err != nil {
		t.Fatalf("erro não esperado: %v", err)
	}
	if contentType != "application/pdf" {
		t.Errorf("contentType esperado application/pdf, obtido %s", contentType)
	}
	if !strings.HasSuffix(filename, ".pdf") {
		t.Errorf("filename deve terminar com .pdf: %s", filename)
	}
	if len(bytes_) < 800 {
		t.Errorf("PDF muito pequeno para ser válido (%d bytes)", len(bytes_))
	}
	// PDF começa com %PDF-
	if !bytes.HasPrefix(bytes_, []byte("%PDF")) {
		t.Errorf("conteúdo não parece PDF válido")
	}
}

func TestGerarRelatorioFormatoInvalido(t *testing.T) {
	inicio := time.Now().Add(-24 * time.Hour)
	fim := time.Now()
	ps := &pontoService{pontoRepo: &mockRepo{registros: nil}}
	_, _, _, err := ps.GerarRelatorio(context.Background(), 10, 1, inicio, fim, "xls")
	if err == nil {
		t.Fatalf("era esperado erro para formato inválido")
	}
}

func TestGerarRelatorioFormatoCaseInsensitive(t *testing.T) {
	inicio := time.Now().Add(-2 * time.Hour)
	fim := time.Now()
	registros := []model.RegistroPonto{{ID: 1, UsuarioID: 10, Timestamp: inicio.Add(30 * time.Minute)}}
	ps := &pontoService{pontoRepo: &mockRepo{registros: registros}}
	// Maiúsculo
	if _, ct, fn, err := ps.GerarRelatorio(context.Background(), 10, 1, inicio, fim, "PDF"); err != nil || ct != "application/pdf" || !strings.HasSuffix(fn, ".pdf") {
		t.Fatalf("esperado PDF válido em formato maiúsculo, err=%v ct=%s fn=%s", err, ct, fn)
	}
	if _, ct, fn, err := ps.GerarRelatorio(context.Background(), 10, 1, inicio, fim, "CSV"); err != nil || ct != "text/csv" || !strings.HasSuffix(fn, ".csv") {
		t.Fatalf("esperado CSV válido em formato maiúsculo, err=%v ct=%s fn=%s", err, ct, fn)
	}
}

func TestGerarRelatorioIntervaloInvertido(t *testing.T) {
	fim := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	inicio := time.Date(2025, 10, 5, 0, 0, 0, 0, time.UTC) // invertido (inicio > fim)
	mr := &mockRepo{registros: nil}
	ps := &pontoService{pontoRepo: mr}
	if _, _, _, err := ps.GerarRelatorio(context.Background(), 10, 1, inicio, fim, "csv"); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if mr.lastInicio.After(mr.lastFim) {
		t.Errorf("intervalo não foi normalizado: inicio=%s fim=%s", mr.lastInicio, mr.lastFim)
	}
	if !mr.lastInicio.Equal(fim) || !mr.lastFim.Equal(inicio) {
		t.Errorf("esperado troca de datas: got inicio=%s fim=%s", mr.lastInicio, mr.lastFim)
	}
}
