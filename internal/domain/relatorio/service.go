package relatorio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/bancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/justificativa"
	"github.com/Loviiin/ponto-api-go/internal/domain/logbancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/cache"
	"github.com/jung-kurt/gofpdf"
)

type Service interface {
	GerarEspelhoPonto(userID uint, empresaID uint, inicio, fim time.Time) (*EspelhoPonto, error)
	GerarRelatorioGeral(dataInicio, dataFim time.Time, usuarioID *uint, empresaID uint, userIDLogado uint) (*RelatorioGeralDTO, error)
	ExportarRelatorioGeral(dataInicio, dataFim time.Time, usuarioID *uint, empresaID uint, userIDLogado uint, formato string, limite, offset int) ([]byte, string, string, error)
}

// Minimal interface to obtain user with contract & cargo
type usuarioReader interface {
	FindByID(ctx context.Context, id uint, empresaID uint) (*model.Usuario, error)
}

type service struct {
	pontoRepo         ponto.RegistroPontoRepository
	usuarioRead       usuarioReader
	logRepo           logbancohoras.Repository
	justificativaRepo justificativa.Repository
	bancoHorasService bancohoras.BancoHorasService
	cache             cache.Service
}

func NewService(pontoRepo ponto.RegistroPontoRepository, usuarioRepo usuarioReader, logRepo logbancohoras.Repository, bancoHorasService bancohoras.BancoHorasService) Service {
	return &service{pontoRepo: pontoRepo, usuarioRead: usuarioRepo, logRepo: logRepo, bancoHorasService: bancoHorasService, justificativaRepo: nil, cache: nil}
}

func NewServiceWithJustificativa(pontoRepo ponto.RegistroPontoRepository, usuarioRepo usuarioReader, logRepo logbancohoras.Repository, justificativaRepo justificativa.Repository, bancoHorasService bancohoras.BancoHorasService) Service {
	return &service{pontoRepo: pontoRepo, usuarioRead: usuarioRepo, logRepo: logRepo, justificativaRepo: justificativaRepo, bancoHorasService: bancoHorasService, cache: nil}
}

func NewServiceWithCache(pontoRepo ponto.RegistroPontoRepository, usuarioRepo usuarioReader, logRepo logbancohoras.Repository, justificativaRepo justificativa.Repository, bancoHorasService bancohoras.BancoHorasService, cacheService cache.Service) Service {
	return &service{pontoRepo: pontoRepo, usuarioRead: usuarioRepo, logRepo: logRepo, justificativaRepo: justificativaRepo, bancoHorasService: bancoHorasService, cache: cacheService}
}

// validarHierarquiaAcesso verifica se o usuário logado tem permissão para acessar os dados do usuário alvo
// Retorna erro se não tiver permissão
func (s *service) validarHierarquiaAcesso(userIDLogado uint, empresaID uint, usuarioAlvoID *uint) error {
	// Se não especificou usuário alvo, é um relatório geral (todos) - precisa ter permissão especial
	// (essa validação é feita pelo middleware, aqui só validamos hierarquia individual)
	if usuarioAlvoID == nil {
		return nil // Middleware já validou VISUALIZAR_RELATORIOS_GERAIS
	}

	// Se está acessando os próprios dados, sempre permitido
	if *usuarioAlvoID == userIDLogado {
		return nil
	}

	// Buscar dados do usuário logado com cargo
	usuarioLogado, err := s.usuarioRead.FindByID(context.Background(), userIDLogado, empresaID)
	if err != nil {
		return fmt.Errorf("erro ao buscar usuário logado: %w", err)
	}

	// Buscar dados do usuário alvo com cargo
	usuarioAlvo, err := s.usuarioRead.FindByID(context.Background(), *usuarioAlvoID, empresaID)
	if err != nil {
		return fmt.Errorf("erro ao buscar usuário alvo: %w", err)
	}

	// Validar hierarquia: usuário logado só pode acessar usuários com nível hierárquico MENOR
	// (quanto maior o número, maior o nível - Dono=100, Gerente=70, Colaborador=10)
	if usuarioLogado.Contrato.Cargo.NivelHierarquia <= usuarioAlvo.Contrato.Cargo.NivelHierarquia {
		return fmt.Errorf("sem permissão para acessar dados de usuário com mesmo nível ou superior na hierarquia")
	}

	return nil
}

func (s *service) GerarEspelhoPonto(userID uint, empresaID uint, inicio, fim time.Time) (*EspelhoPonto, error) {
	if fim.Before(inicio) {
		inicio, fim = fim, inicio
	}

	usr, err := s.usuarioRead.FindByID(context.Background(), userID, empresaID)
	if err != nil {
		return nil, err
	}
	if usr.Contrato.ID == 0 || usr.Contrato.Cargo.ID == 0 {
		return nil, errors.New("usuário sem contrato/cargo para calcular espelho")
	}

	registros, err := s.pontoRepo.FindPontosByUserIDAndDateRange(userID, inicio, fim)
	if err != nil {
		return nil, err
	}

	// Agrupar por dia
	porDia := map[string][]model.RegistroPonto{}
	for _, r := range registros {
		key := r.Timestamp.Format("2006-01-02")
		porDia[key] = append(porDia[key], r)
	}

	// Obter lista de dias no intervalo
	diasLista := []string{}
	for d := inicio; !d.After(fim); d = d.Add(24 * time.Hour) {
		diasLista = append(diasLista, d.Format("2006-01-02"))
	}

	cargaDia := int(usr.Contrato.Cargo.CargaHorariaDiariaMinutos)
	saldoAcumulado := 0
	var diasEspelho []DiaEspelho
	var diasInconsistentes []string
	trabalhadoTotal := 0

	for _, diaStr := range diasLista {
		entries := porDia[diaStr]
		sort.Slice(entries, func(i, j int) bool { return entries[i].Timestamp.Before(entries[j].Timestamp) })
		inconsistente := len(entries)%2 == 1
		if inconsistente {
			diasInconsistentes = append(diasInconsistentes, diaStr)
		}

		// Pairing
		var pares []ParMarcacao
		var totalMinutos int
		for i := 0; i+1 < len(entries); i += 2 {
			entrada := entries[i].Timestamp
			saida := entries[i+1].Timestamp
			dur := int(saida.Sub(entrada).Minutes())
			totalMinutos += dur
			entradaCopy := entrada
			saidaCopy := saida
			pares = append(pares, ParMarcacao{Entrada: &entradaCopy, Saida: &saidaCopy})
		}
		trabalhadoTotal += totalMinutos
		// Reutiliza a lógica do serviço de banco de horas para calcular o saldo do dia
		// Convertendo diaStr de volta para time.Time
		diaTime, _ := time.Parse("2006-01-02", diaStr)
		saldoDia, errSaldo := s.bancoHorasService.CalcularSaldoParaUsuario(usr.ID, usr.Contrato.EmpresaID, diaTime)
		if errSaldo != nil {
			// fallback para cálculo local caso dê erro (ex.: marcações ímpares)
			saldoDia = totalMinutos - cargaDia
		}
		saldoAcumulado += saldoDia

		// Fechado (heurística: existe log com motivo contendo a data)
		fechado := false
		logMotivo := ""
		// (Eficiência: para um MVP não buscamos logs ainda; será adicionado depois com repositório extendido)

		diasEspelho = append(diasEspelho, DiaEspelho{
			Data:                   diaStr,
			Marcacoes:              pares,
			TotalTrabalhadoMinutos: totalMinutos,
			TotalTrabalhadoHHMM:    formatHHMM(totalMinutos),
			CargaPlanejadaMinutos:  cargaDia,
			SaldoDiaMinutos:        saldoDia,
			SaldoAcumuladoMinutos:  saldoAcumulado,
			Fechado:                fechado,
			LogMotivo:              logMotivo,
			Inconsistente:          inconsistente,
		})
	}

	espelho := &EspelhoPonto{
		UsuarioID:          userID,
		Periodo:            PeriodoInterval{Inicio: inicio.Format("2006-01-02"), Fim: fim.Format("2006-01-02")},
		Dias:               diasEspelho,
		Totais:             TotaisPeriodo{TrabalhadoMinutos: trabalhadoTotal, TrabalhadoHHMM: formatHHMM(trabalhadoTotal), SaldoFinalMinutos: saldoAcumulado},
		DiasInconsistentes: diasInconsistentes,
	}
	return espelho, nil
}

func formatHHMM(total int) string {
	h := total / 60
	m := total % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

// GerarRelatorioGeral gera um relatório consolidado para um ou mais usuários em um período
func (s *service) GerarRelatorioGeral(dataInicio, dataFim time.Time, usuarioID *uint, empresaID uint, userIDLogado uint) (*RelatorioGeralDTO, error) {
	if dataFim.Before(dataInicio) {
		dataInicio, dataFim = dataFim, dataInicio
	}

	// Validar hierarquia de acesso
	if err := s.validarHierarquiaAcesso(userIDLogado, empresaID, usuarioID); err != nil {
		return nil, err
	}

	// Tentar obter do cache
	cacheKey := gerarCacheKeyRelatorio(empresaID, usuarioID, dataInicio, dataFim)
	if s.cache != nil {
		if cached, err := s.cache.Get(context.Background(), cacheKey); err == nil && cached != "" {
			var relatorio RelatorioGeralDTO
			if errUM := json.Unmarshal([]byte(cached), &relatorio); errUM == nil {
				log.Printf("[cache] HIT %s", cacheKey)
				return &relatorio, nil
			}
		}
	}

	// Lista de usuários a processar
	var usuarios []model.Usuario
	var err error

	if usuarioID != nil {
		// Buscar usuário específico
		usr, errU := s.usuarioRead.FindByID(context.Background(), *usuarioID, empresaID)
		if errU != nil {
			return nil, fmt.Errorf("usuário não encontrado: %w", errU)
		}
		usuarios = append(usuarios, *usr)
	} else {
		// Buscar todos os usuários ativos da empresa
		// Precisamos de um método GetAllActive no repositório
		if usuarioRepo, ok := s.usuarioRead.(interface {
			GetAllActive(empresaID uint) ([]model.Usuario, error)
		}); ok {
			usuarios, err = usuarioRepo.GetAllActive(empresaID)
			if err != nil {
				return nil, fmt.Errorf("erro ao buscar usuários ativos: %w", err)
			}
		} else {
			return nil, errors.New("repositório não suporta GetAllActive")
		}
	}

	if len(usuarios) == 0 {
		return nil, errors.New("nenhum usuário encontrado para gerar relatório")
	}

	// Gerar relatório para cada usuário
	var relatoriosUsuarios []RelatorioUsuarioDTO

	for _, usr := range usuarios {
		if usr.Contrato.ID == 0 || usr.Contrato.Cargo.ID == 0 {
			// Usuário sem contrato ou cargo, pular
			continue
		}

		relatorioUsuario, err := s.gerarRelatorioParaUsuario(usr, empresaID, dataInicio, dataFim)
		if err != nil {
			// Log do erro mas continua processando outros usuários
			fmt.Printf("Erro ao gerar relatório para usuário %d: %v\n", usr.ID, err)
			continue
		}

		relatoriosUsuarios = append(relatoriosUsuarios, *relatorioUsuario)
	}

	relatorioGeral := &RelatorioGeralDTO{
		Periodo: PeriodoInterval{
			Inicio: dataInicio.Format("2006-01-02"),
			Fim:    dataFim.Format("2006-01-02"),
		},
		Usuarios: relatoriosUsuarios,
	}

	// Salvar no cache (15 minutos para relatórios dinâmicos)
	if s.cache != nil {
		if b, mErr := json.Marshal(relatorioGeral); mErr == nil {
			ttl := 15 * time.Minute
			if sErr := s.cache.Set(context.Background(), cacheKey, string(b), ttl); sErr != nil {
				log.Printf("[cache] erro ao setar chave %s: %v", cacheKey, sErr)
			} else {
				log.Printf("[cache] SET %s (TTL: %v)", cacheKey, ttl)
			}
		}
	}

	return relatorioGeral, nil
}

// gerarRelatorioParaUsuario gera o relatório detalhado para um único usuário
func (s *service) gerarRelatorioParaUsuario(usr model.Usuario, empresaID uint, dataInicio, dataFim time.Time) (*RelatorioUsuarioDTO, error) {
	// Buscar registros de ponto no período
	registros, err := s.pontoRepo.FindPontosByUserIDAndDateRange(usr.ID, dataInicio, dataFim)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar pontos: %w", err)
	}

	// Buscar justificativas no período (apenas aprovadas)
	var justificativasMap map[string]*model.Justificativa
	if s.justificativaRepo != nil {
		justificativas, err := s.justificativaRepo.FindByUsuarioIDAndPeriodo(usr.ID, empresaID, dataInicio, dataFim)
		if err == nil {
			justificativasMap = make(map[string]*model.Justificativa)
			for i := range justificativas {
				if justificativas[i].Status == "APROVADA" {
					dataKey := justificativas[i].DataOcorrencia.Format("2006-01-02")
					justificativasMap[dataKey] = &justificativas[i]
				}
			}
		}
	} else {
		justificativasMap = make(map[string]*model.Justificativa)
	}

	// Agrupar registros por dia
	porDia := map[string][]model.RegistroPonto{}
	for _, r := range registros {
		key := r.Timestamp.Format("2006-01-02")
		porDia[key] = append(porDia[key], r)
	}

	// Obter lista de dias no intervalo
	diasLista := []string{}
	for d := dataInicio; !d.After(dataFim); d = d.Add(24 * time.Hour) {
		diasLista = append(diasLista, d.Format("2006-01-02"))
	}

	cargaDiariaMinutos := int(usr.Contrato.Cargo.CargaHorariaDiariaMinutos)

	// Totalizadores
	totalTrabalhadoMinutos := 0
	totalExtrasMinutos := 0
	totalFaltantesMinutos := 0
	saldoBancoHorasInicial := usr.Contrato.SaldoBancoHorasMinutos
	saldoBancoHorasAtual := saldoBancoHorasInicial

	var diasRelatorio []RelatorioDiaDTO

	for _, diaStr := range diasLista {
		entries := porDia[diaStr]
		sort.Slice(entries, func(i, j int) bool { return entries[i].Timestamp.Before(entries[j].Timestamp) })

		inconsistente := len(entries)%2 == 1

		// Parear entradas e saídas
		var pares []ParMarcacao
		var totalMinutosDia int
		for i := 0; i+1 < len(entries); i += 2 {
			entrada := entries[i].Timestamp
			saida := entries[i+1].Timestamp
			dur := int(saida.Sub(entrada).Minutes())
			totalMinutosDia += dur
			entradaCopy := entrada
			saidaCopy := saida
			pares = append(pares, ParMarcacao{Entrada: &entradaCopy, Saida: &saidaCopy})
		}

		totalTrabalhadoMinutos += totalMinutosDia

		// Calcular horas extras/faltantes
		var horasExtras, horasFaltantes int
		diferenca := totalMinutosDia - cargaDiariaMinutos

		// Se há justificativa aprovada, pode abonar as horas faltantes
		justificativa := justificativasMap[diaStr]
		if justificativa != nil {
			// Com justificativa aprovada, considera que cumpriu a carga
			if diferenca < 0 {
				// Tinha horas faltantes mas foi justificado
				horasFaltantes = 0
			} else {
				horasExtras = diferenca
			}
		} else {
			// Sem justificativa
			if diferenca > 0 {
				horasExtras = diferenca
			} else if diferenca < 0 {
				horasFaltantes = -diferenca
			}
		}

		totalExtrasMinutos += horasExtras
		totalFaltantesMinutos += horasFaltantes

		// Atualizar saldo do banco de horas
		saldoBancoHorasAtual += diferenca

		// Montar DTO de justificativa se houver
		var justificativaDTO *JustificativaDTO
		if justificativa != nil {
			justificativaDTO = &JustificativaDTO{
				ID:        justificativa.ID,
				Tipo:      justificativa.Tipo,
				Descricao: justificativa.Descricao,
				Status:    justificativa.Status,
			}
		}

		diaDT := RelatorioDiaDTO{
			Data:                   diaStr,
			Marcacoes:              pares,
			TotalTrabalhadoMinutos: totalMinutosDia,
			TotalTrabalhadoHHMM:    formatHHMM(totalMinutosDia),
			CargaEsperadaMinutos:   cargaDiariaMinutos,
			CargaEsperadaHHMM:      formatHHMM(cargaDiariaMinutos),
			HorasExtrasMinutos:     horasExtras,
			HorasFaltantesMinutos:  horasFaltantes,
			SaldoBancoHorasMinutos: saldoBancoHorasAtual,
			Justificativa:          justificativaDTO,
			Inconsistente:          inconsistente,
		}

		diasRelatorio = append(diasRelatorio, diaDT)
	}

	// Montar resumo
	resumo := ResumoUsuarioDTO{
		TotalHorasTrabalhadasMinutos:  totalTrabalhadoMinutos,
		TotalHorasTrabalhadasHHMM:     formatHHMM(totalTrabalhadoMinutos),
		TotalHorasExtrasMinutos:       totalExtrasMinutos,
		TotalHorasExtrasHHMM:          formatHHMM(totalExtrasMinutos),
		TotalHorasFaltantesMinutos:    totalFaltantesMinutos,
		TotalHorasFaltantesHHMM:       formatHHMM(totalFaltantesMinutos),
		SaldoBancoHorasInicialMinutos: saldoBancoHorasInicial,
		SaldoBancoHorasFinalMinutos:   saldoBancoHorasAtual,
	}

	// Montar informações do usuário
	usuarioInfo := UsuarioInfoDTO{
		ID:        usr.ID,
		Nome:      usr.Nome,
		CPF:       usr.CPF,
		Email:     usr.Email,
		CargoNome: usr.Contrato.Cargo.Nome,
	}

	relatorioUsuario := &RelatorioUsuarioDTO{
		Usuario: usuarioInfo,
		Periodo: PeriodoInterval{
			Inicio: dataInicio.Format("2006-01-02"),
			Fim:    dataFim.Format("2006-01-02"),
		},
		Dias:   diasRelatorio,
		Resumo: resumo,
	}

	return relatorioUsuario, nil
}

var ErrFormatoInvalido = errors.New("formato inválido; use 'csv' ou 'pdf'")

// ExportarRelatorioGeral gera um arquivo (CSV ou PDF) do relatório geral
func (s *service) ExportarRelatorioGeral(dataInicio, dataFim time.Time, usuarioID *uint, empresaID uint, userIDLogado uint, formato string, limite, offset int) ([]byte, string, string, error) {
	if dataInicio.After(dataFim) {
		dataInicio, dataFim = dataFim, dataInicio
	}
	formato = strings.ToLower(strings.TrimSpace(formato))

	// Validar hierarquia de acesso
	if err := s.validarHierarquiaAcesso(userIDLogado, empresaID, usuarioID); err != nil {
		return nil, "", "", err
	}

	// Modificar a lógica para aplicar paginação quando usuarioID é nil
	var relatorio *RelatorioGeralDTO
	var err error

	if usuarioID != nil {
		// Relatório de usuário específico - sem paginação
		relatorio, err = s.GerarRelatorioGeral(dataInicio, dataFim, usuarioID, empresaID, userIDLogado)
	} else {
		// Relatório de todos - aplicar paginação
		relatorio, err = s.gerarRelatorioGeralComPaginacao(dataInicio, dataFim, empresaID, limite, offset)
	}

	if err != nil {
		return nil, "", "", err
	}

	switch formato {
	case "csv":
		return gerarRelatorioGeralCSV(relatorio, dataInicio, dataFim)
	case "pdf":
		return gerarRelatorioGeralPDF(relatorio, dataInicio, dataFim)
	default:
		return nil, "", "", ErrFormatoInvalido
	}
}

// gerarRelatorioGeralComPaginacao gera relatório com paginação de usuários
func (s *service) gerarRelatorioGeralComPaginacao(dataInicio, dataFim time.Time, empresaID uint, limite, offset int) (*RelatorioGeralDTO, error) {
	// Buscar usuários com paginação
	if usuarioRepo, ok := s.usuarioRead.(interface {
		GetAllActive(empresaID uint) ([]model.Usuario, error)
	}); ok {
		todosUsuarios, err := usuarioRepo.GetAllActive(empresaID)
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar usuários ativos: %w", err)
		}

		// Aplicar paginação manualmente
		total := len(todosUsuarios)
		if offset >= total {
			return &RelatorioGeralDTO{
				Periodo: PeriodoInterval{
					Inicio: dataInicio.Format("2006-01-02"),
					Fim:    dataFim.Format("2006-01-02"),
				},
				Usuarios: []RelatorioUsuarioDTO{},
			}, nil
		}

		fim := offset + limite
		if fim > total {
			fim = total
		}
		usuariosPaginados := todosUsuarios[offset:fim]

		// Gerar relatório para usuários paginados
		var relatoriosUsuarios []RelatorioUsuarioDTO
		for _, usr := range usuariosPaginados {
			if usr.Contrato.ID == 0 || usr.Contrato.Cargo.ID == 0 {
				continue
			}

			relatorioUsuario, err := s.gerarRelatorioParaUsuario(usr, empresaID, dataInicio, dataFim)
			if err != nil {
				fmt.Printf("Erro ao gerar relatório para usuário %d: %v\n", usr.ID, err)
				continue
			}
			relatoriosUsuarios = append(relatoriosUsuarios, *relatorioUsuario)
		}

		return &RelatorioGeralDTO{
			Periodo: PeriodoInterval{
				Inicio: dataInicio.Format("2006-01-02"),
				Fim:    dataFim.Format("2006-01-02"),
			},
			Usuarios: relatoriosUsuarios,
		}, nil
	}

	return nil, errors.New("repositório não suporta GetAllActive")
}

// gerarRelatorioGeralCSV exporta o relatório geral em formato CSV
func gerarRelatorioGeralCSV(relatorio *RelatorioGeralDTO, inicio, fim time.Time) ([]byte, string, string, error) {
	var b strings.Builder

	// Cabeçalho
	b.WriteString("usuario_id,usuario_nome,usuario_cpf,usuario_cargo,data,hora_entrada,hora_saida,total_trabalhado_hhmm,carga_esperada_hhmm,horas_extras_min,horas_faltantes_min,saldo_banco_horas_min,justificativa_tipo,inconsistente\n")

	for _, usuarioRel := range relatorio.Usuarios {
		for _, dia := range usuarioRel.Dias {
			// Para cada par de marcações no dia
			if len(dia.Marcacoes) == 0 {
				// Dia sem marcações
				b.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s,,,00:00,%s,%d,%d,%d,,false\n",
					usuarioRel.Usuario.ID,
					escaparCSV(usuarioRel.Usuario.Nome),
					usuarioRel.Usuario.CPF,
					escaparCSV(usuarioRel.Usuario.CargoNome),
					dia.Data,
					dia.CargaEsperadaHHMM,
					dia.HorasExtrasMinutos,
					dia.HorasFaltantesMinutos,
					dia.SaldoBancoHorasMinutos,
				))
			} else {
				// Tem marcações - criar uma linha por par
				for _, marc := range dia.Marcacoes {
					entradaStr := ""
					saidaStr := ""
					if marc.Entrada != nil {
						entradaStr = marc.Entrada.Format("15:04:05")
					}
					if marc.Saida != nil {
						saidaStr = marc.Saida.Format("15:04:05")
					}

					justifTipo := ""
					if dia.Justificativa != nil {
						justifTipo = dia.Justificativa.Tipo
					}

					b.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%s,%d,%d,%d,%s,%v\n",
						usuarioRel.Usuario.ID,
						escaparCSV(usuarioRel.Usuario.Nome),
						usuarioRel.Usuario.CPF,
						escaparCSV(usuarioRel.Usuario.CargoNome),
						dia.Data,
						entradaStr,
						saidaStr,
						dia.TotalTrabalhadoHHMM,
						dia.CargaEsperadaHHMM,
						dia.HorasExtrasMinutos,
						dia.HorasFaltantesMinutos,
						dia.SaldoBancoHorasMinutos,
						justifTipo,
						dia.Inconsistente,
					))
				}
			}
		}
	}

	filename := fmt.Sprintf("relatorio_geral_%s_%s.csv", inicio.Format("20060102"), fim.Format("20060102"))
	return []byte(b.String()), "text/csv", filename, nil
}

// gerarRelatorioGeralPDF exporta o relatório geral em formato PDF
func gerarRelatorioGeralPDF(relatorio *RelatorioGeralDTO, inicio, fim time.Time) ([]byte, string, string, error) {
	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape para mais espaço
	pdf.SetMargins(10, 15, 10)
	pdf.SetAutoPageBreak(true, 15)

	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local
	}

	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Footer com página
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("Arial", "", 8)
		pdf.CellFormat(0, 8, fmt.Sprintf("Página %d", pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	pdf.AddPage()

	// Cabeçalho geral
	pdf.SetFont("Arial", "B", 16)
	titulo := "Relatório Geral de Ponto"
	pdf.CellFormat(0, 10, tr(titulo), "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	periodo := fmt.Sprintf("Período: %s a %s", inicio.In(loc).Format("02/01/2006"), fim.In(loc).Format("02/01/2006"))
	geradoEm := fmt.Sprintf("Gerado em: %s", time.Now().In(loc).Format("02/01/2006 15:04:05"))
	pdf.CellFormat(0, 6, tr(periodo), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, tr(geradoEm), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("Total de usuários: %d", len(relatorio.Usuarios)), "", 1, "L", false, 0, "")
	pdf.Ln(3)

	// Para cada usuário
	for _, usuarioRel := range relatorio.Usuarios {
		// Nova página para cada usuário se necessário
		if pdf.GetY() > 180 {
			pdf.AddPage()
		}

		// Cabeçalho do usuário
		pdf.SetFillColor(240, 240, 240)
		pdf.SetFont("Arial", "B", 12)
		nomeUsuario := fmt.Sprintf("%s (ID: %d) - %s", usuarioRel.Usuario.Nome, usuarioRel.Usuario.ID, usuarioRel.Usuario.CargoNome)
		pdf.CellFormat(0, 8, tr(nomeUsuario), "1", 1, "L", true, 0, "")

		// Resumo do usuário
		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(0, 5, fmt.Sprintf("Total Trabalhado: %s | Extras: %s | Faltantes: %s | Saldo Banco Horas: %d min",
			usuarioRel.Resumo.TotalHorasTrabalhadasHHMM,
			usuarioRel.Resumo.TotalHorasExtrasHHMM,
			usuarioRel.Resumo.TotalHorasFaltantesHHMM,
			usuarioRel.Resumo.SaldoBancoHorasFinalMinutos,
		), "1", 1, "L", false, 0, "")
		pdf.Ln(2)

		// Tabela de dias (simplificada)
		headers := []string{"Data", "Trab.", "Extras", "Falt.", "Saldo BH"}
		colWidths := []float64{35, 20, 20, 20, 25}

		// Cabeçalho da tabela
		pdf.SetFillColor(30, 60, 110)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Arial", "B", 8)
		for i, h := range headers {
			pdf.CellFormat(colWidths[i], 6, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Dados
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("Arial", "", 8)
		fill := false
		for _, dia := range usuarioRel.Dias {
			if fill {
				pdf.SetFillColor(245, 245, 245)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}

			pdf.CellFormat(colWidths[0], 5, dia.Data, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(colWidths[1], 5, dia.TotalTrabalhadoHHMM, "1", 0, "C", fill, 0, "")
			pdf.CellFormat(colWidths[2], 5, fmt.Sprintf("%d min", dia.HorasExtrasMinutos), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(colWidths[3], 5, fmt.Sprintf("%d min", dia.HorasFaltantesMinutos), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(colWidths[4], 5, fmt.Sprintf("%d min", dia.SaldoBancoHorasMinutos), "1", 0, "C", fill, 0, "")
			pdf.Ln(-1)
			fill = !fill
		}

		pdf.Ln(5)
	}

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, "", "", err
	}

	filename := fmt.Sprintf("relatorio_geral_%s_%s.pdf", inicio.Format("20060102"), fim.Format("20060102"))
	return buf.Bytes(), "application/pdf", filename, nil
}

// escaparCSV escapa strings para CSV
func escaparCSV(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}



// gerarCacheKeyRelatorio gera uma chave única para cache de relatórios
func gerarCacheKeyRelatorio(empresaID uint, usuarioID *uint, dataInicio, dataFim time.Time) string {
	userPart := "todos"
	if usuarioID != nil {
		userPart = fmt.Sprintf("u%d", *usuarioID)
	}
	return fmt.Sprintf("relatorio_geral:e%d:%s:%s:%s",
		empresaID,
		userPart,
		dataInicio.Format("20060102"),
		dataFim.Format("20060102"),
	)
}
