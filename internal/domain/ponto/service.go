package ponto

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/domain/localidade"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/jung-kurt/gofpdf"
	"github.com/umahmood/haversine"
	"gorm.io/gorm"
)

type PontoService interface {
	BaterPonto(usuarioID uint, empresaID uint, latitude, longitude float64) (*model.RegistroPonto, error)
	GetPontosDoDia(usuarioID uint, dia time.Time) ([]model.RegistroPonto, error)
	AjustarPonto(usuarioID, empresaID, adminID uint, timestamp time.Time, justificativaID *uint) (*model.RegistroPonto, error)
	EditarPonto(pontoID, empresaID uint, novoTimestamp time.Time, justificativaID *uint) (*model.RegistroPonto, error)
	FindPontoByID(pontoID, empresaID uint) (*model.RegistroPonto, error)
	GerarRelatorio(userID, empresaID uint, inicio, fim time.Time, formato string) ([]byte, string, string, error)
}

var ErrFormatoInvalido = errors.New("formato inválido; use 'csv' ou 'pdf'")
var ErrGeofencingViolation = errors.New("geofencing violation")

type pontoService struct {
	pontoRepo      RegistroPontoRepository
	userRepo       usuario.UsuarioRepository
	localidadeRepo localidade.Repository
	empresaRepo    empresa.EmpresaRepository
	db             *gorm.DB
}

func NewPontoService(
	pontoRepo RegistroPontoRepository,
	userRepo usuario.UsuarioRepository,
	localidadeRepo localidade.Repository,
	empresaRepo empresa.EmpresaRepository,
	db *gorm.DB,
) PontoService {
	return &pontoService{
		pontoRepo:      pontoRepo,
		userRepo:       userRepo,
		localidadeRepo: localidadeRepo,
		empresaRepo:    empresaRepo,
		db:             db,
	}
}

func (s *pontoService) BaterPonto(usuarioID uint, empresaID uint, latitude, longitude float64) (*model.RegistroPonto, error) {
	user, err := s.userRepo.FindByID(context.Background(), usuarioID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usuário não encontrado ou não pertence a esta empresa")
		}
		return nil, err
	}

	if user.Contrato.ID == 0 {
		return nil, errors.New("usuário não possui um contrato de trabalho ativo")
	}

	localidade, err := s.localidadeRepo.FindByID(user.Contrato.LocalidadeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("a localidade de trabalho do usuário não foi encontrada")
		}
		return nil, err
	}

	pontoLocalidade := haversine.Coord{Lat: localidade.Latitude, Lon: localidade.Longitude}
	pontoBatida := haversine.Coord{Lat: latitude, Lon: longitude}

	km, _ := haversine.Distance(pontoLocalidade, pontoBatida)
	distanciaEmMetros := km * 1000

	var tipoBatida string
	dentroDoRaio := distanciaEmMetros <= localidade.RaioGeofenceMetros
	
	if dentroDoRaio {
		tipoBatida = "Presencial"
	} else {
		tipoBatida = "Remoto"
	}

	// Verificar se a empresa requer que batidas presenciais sejam dentro do raio
	empresa, err := s.empresaRepo.FindByID(context.Background(), empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("empresa não encontrada")
		}
		return nil, err
	}

	// Se a restrição está ativa e o usuário está fora do raio, rejeitar a batida
	// (independente da tentativa ser presencial ou remota - a regra bloqueia qualquer batida fora do raio)
	if empresa.PresencialRestritoAoRaio && !dentroDoRaio {
		return nil, fmt.Errorf("%w: você está fora do raio permitido (%.2fm de distância, raio máximo: %.2fm)", ErrGeofencingViolation, distanciaEmMetros, localidade.RaioGeofenceMetros)
	}

	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local
	}

	registroPonto := &model.RegistroPonto{
		UsuarioID: usuarioID,
		Latitude:  latitude,
		Longitude: longitude,
		Timestamp: time.Now().In(loc),
		EmpresaID: empresaID,
		Metodo:    tipoBatida,
	}

	if err := s.pontoRepo.SavePonto(registroPonto); err != nil {
		return nil, err
	}

	return registroPonto, nil
}

func (s *pontoService) GetPontosDoDia(usuarioID uint, dia time.Time) ([]model.RegistroPonto, error) {
	return s.pontoRepo.FindPontosByUserIDAndDate(usuarioID, dia)
}

func (s *pontoService) AjustarPonto(usuarioID, empresaID, adminID uint, timestamp time.Time, justificativaID *uint) (*model.RegistroPonto, error) {
	_, err := s.userRepo.FindByID(context.Background(), usuarioID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usuário alvo não encontrado ou não pertence a esta empresa")
		}
		return nil, err
	}

	pontoRegistrado := &model.RegistroPonto{
		UsuarioID:       usuarioID,
		EmpresaID:       empresaID,
		Timestamp:       timestamp,
		Metodo:          "AJUSTE_MANUAL_ADMIN",
		Localizacao:     "N/A",
		JustificativaID: justificativaID,
	}

	if err := s.pontoRepo.SavePonto(pontoRegistrado); err != nil {
		return nil, err
	}

	return pontoRegistrado, nil
}

func (s *pontoService) EditarPonto(pontoID, empresaID uint, novoTimestamp time.Time, justificativaID *uint) (*model.RegistroPonto, error) {
	pontoParaEditar, err := s.pontoRepo.FindPontoByID(pontoID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("registro de ponto não encontrado ou não pertence a esta empresa")
		}
		return nil, err
	}

	pontoParaEditar.Timestamp = novoTimestamp
	pontoParaEditar.Metodo = "EDICAO_MANUAL_ADMIN"
	pontoParaEditar.JustificativaID = justificativaID

	if err := s.pontoRepo.UpdatePonto(pontoParaEditar); err != nil {
		return nil, err
	}

	return pontoParaEditar, nil
}

func (s *pontoService) FindPontoByID(pontoID, empresaID uint) (*model.RegistroPonto, error) {
	return s.pontoRepo.FindPontoByID(pontoID, empresaID)
}

func (s *pontoService) GerarRelatorio(userID, empresaID uint, inicio, fim time.Time, formato string) ([]byte, string, string, error) {
	if inicio.After(fim) {
		inicio, fim = fim, inicio
	}
	formato = strings.ToLower(strings.TrimSpace(formato))

	// Buscar usuário para obter nome (se existir) - guarda nil para testes
	userName := fmt.Sprintf("Usuário %d", userID)
	if s.userRepo != nil {
		if usuarioObj, errUser := s.userRepo.FindByID(context.Background(), userID, empresaID); errUser == nil && usuarioObj != nil && usuarioObj.Nome != "" {
			userName = usuarioObj.Nome
		}
	}
	registros, err := s.pontoRepo.FindPontosByUserIDAndDateRange(userID, inicio, fim)
	if err != nil {
		return nil, "", "", err
	}

	switch formato {
	case "csv":
		return gerarRelatorioCSV(registros, userID, userName, inicio, fim)
	case "pdf":
		return gerarRelatorioPDF(registros, userID, userName, inicio, fim)
	default:
		return nil, "", "", ErrFormatoInvalido
	}
}

func gerarRelatorioCSV(registros []model.RegistroPonto, userID uint, userName string, inicio, fim time.Time) ([]byte, string, string, error) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local
	}
	var b strings.Builder
	// Cabeçalho
	b.WriteString("id,timestamp,latitude,longitude,metodo,justificativa_id\n")
	for _, r := range registros {
		b.WriteString(
			fmt.Sprintf("%d,%s,%.6f,%.6f,%s,", r.ID, r.Timestamp.In(loc).Format(time.RFC3339), r.Latitude, r.Longitude, r.Metodo),
		)
		if r.JustificativaID != nil {
			b.WriteString(fmt.Sprintf("%d", *r.JustificativaID))
		}
		b.WriteString("\n")
	}
	slug := gerarSlugNome(userName)
	filename := fmt.Sprintf("relatorio_ponto_%s_%d_%s_%s.csv", slug, userID, inicio.Format("20060102"), fim.Format("20060102"))
	return []byte(b.String()), "text/csv", filename, nil
}

func gerarRelatorioPDF(registros []model.RegistroPonto, userID uint, userName string, inicio, fim time.Time) ([]byte, string, string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 15, 12)
	pdf.SetAutoPageBreak(true, 15)

	// Timezone Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local
	}

	hasDejaVu := false
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// (Opcional futuramente) tentar registrar fonte somente se realmente necessário e o arquivo for válido.
	// Mantido desativado para evitar o 500.
	// if len(fontDejaVu) > 0 {
	// 	pdf.AddUTF8FontFromBytes("DejaVu", "", fontDejaVu)
	// 	if pdf.Error() == nil { // somente considera se não houve erro interno
	// 		hasDejaVu = true
	// 		tr = func(s string) string { return s }
	// 	} else {
	// 		// Limpa erro para não abortar o PDF
	// 		_ = pdf.Error()
	// 		pdf.SetError(nil)
	// 	}
	// }

	// Footer com página
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("Arial", "", 8)
		pdf.CellFormat(0, 8, fmt.Sprintf("Página %d", pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	pdf.AddPage()

	// Cabeçalho
	if hasDejaVu {
		pdf.SetFont("DejaVu", "", 15)
	} else {
		pdf.SetFont("Arial", "B", 15)
	}
	titulo := fmt.Sprintf("Relatório de Ponto - %s (ID %d)", userName, userID)
	pdf.CellFormat(0, 10, tr(titulo), "", 1, "C", false, 0, "")
	if hasDejaVu {
		pdf.SetFont("DejaVu", "", 10)
	} else {
		pdf.SetFont("Arial", "", 10)
	}
	periodo := fmt.Sprintf("Período: %s a %s", inicio.In(loc).Format("02/01/2006"), fim.In(loc).Format("02/01/2006"))
	geradoEm := fmt.Sprintf("Gerado em: %s", time.Now().In(loc).Format("02/01/2006 15:04:05"))
	pdf.CellFormat(0, 6, tr(periodo), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, tr(geradoEm), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	headers := []string{"ID", "Timestamp", "Latitude", "Longitude", "Método", "Justif."}
	colWidths := []float64{12, 40, 25, 25, 32, 20}

	// Função para desenhar cabeçalho da tabela
	drawHeader := func() {
		pdf.SetFillColor(30, 60, 110)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetDrawColor(200, 200, 200)
		pdf.SetLineWidth(0.2)
		if hasDejaVu {
			pdf.SetFont("DejaVu", "", 9)
		} else {
			pdf.SetFont("Arial", "B", 9)
		}
		for i, h := range headers {
			pdf.CellFormat(colWidths[i], 7, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)
		// Reset para linhas
		if hasDejaVu {
			pdf.SetFont("DejaVu", "", 8)
		} else {
			pdf.SetFont("Arial", "", 8)
		}
		pdf.SetTextColor(0, 0, 0)
	}

	drawHeader()

	// Linhas
	fill := false
	total := len(registros)
	for _, r := range registros {
		if pdf.GetY() > 260 { // próxima página se perto do fim
			pdf.AddPage()
			drawHeader()
		}
		if fill {
			pdf.SetFillColor(240, 244, 248)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		fill = !fill

		just := ""
		if r.JustificativaID != nil {
			just = fmt.Sprintf("%d", *r.JustificativaID)
		}
		row := []string{
			fmt.Sprintf("%d", r.ID),
			r.Timestamp.In(loc).Format("02/01/2006 15:04:05"),
			fmt.Sprintf("%.5f", r.Latitude),
			fmt.Sprintf("%.5f", r.Longitude),
			r.Metodo,
			just,
		}
		for i, val := range row {
			pdf.CellFormat(colWidths[i], 6, tr(val), "1", 0, "L", true, 0, "")
		}
		pdf.Ln(-1)
	}

	// Resumo
	pdf.Ln(2)
	if hasDejaVu {
		pdf.SetFont("DejaVu", "", 8)
	} else {
		pdf.SetFont("Arial", "", 8)
	}
	pdf.CellFormat(0, 6, tr(fmt.Sprintf("Total de registros: %d", total)), "", 1, "L", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", "", err
	}
	slug := gerarSlugNome(userName)
	filename := fmt.Sprintf("relatorio_ponto_%s_%d_%s_%s.pdf", slug, userID, inicio.Format("20060102"), fim.Format("20060102"))
	return buf.Bytes(), "application/pdf", filename, nil
}

// gerarSlugNome cria uma versão segura do nome para uso em filenames.
func gerarSlugNome(nome string) string {
	nome = strings.ToLower(strings.TrimSpace(nome))
	// remover acentos simples substituindo por equivalente ASCII básico
	// abordagem simples: trocar caracteres comuns manualmente
	repl := map[string]string{
		"á": "a", "à": "a", "â": "a", "ã": "a", "ä": "a",
		"é": "e", "è": "e", "ê": "e", "ë": "e",
		"í": "i", "ì": "i", "î": "i", "ï": "i",
		"ó": "o", "ò": "o", "ô": "o", "õ": "o", "ö": "o",
		"ú": "u", "ù": "u", "û": "u", "ü": "u",
		"ç": "c",
	}
	for k, v := range repl {
		nome = strings.ReplaceAll(nome, k, v)
	}
	// manter apenas letras, numeros e converter espaços / separadores em '-'
	espacos := regexp.MustCompile(`[\s_]+`)
	nome = espacos.ReplaceAllString(nome, "-")
	inval := regexp.MustCompile(`[^a-z0-9\-]`)
	nome = inval.ReplaceAllString(nome, "")
	if nome == "" {
		return "usuario"
	}
	if len(nome) > 40 {
		nome = nome[:40]
	}
	return nome
}
