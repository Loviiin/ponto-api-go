package relatorio

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/logbancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	"github.com/Loviiin/ponto-api-go/internal/model"
)

type Service interface {
	GerarEspelhoPonto(userID uint, empresaID uint, inicio, fim time.Time) (*EspelhoPonto, error)
}

// Minimal interface to obtain user with contract & cargo
type usuarioReader interface {
	FindByID(id uint, empresaID uint) (*model.Usuario, error)
}

type service struct {
	pontoRepo   ponto.RegistroPontoRepository
	usuarioRead usuarioReader
	logRepo     logbancohoras.Repository
}

func NewService(pontoRepo ponto.RegistroPontoRepository, usuarioRepo usuarioReader, logRepo logbancohoras.Repository) Service {
	return &service{pontoRepo: pontoRepo, usuarioRead: usuarioRepo, logRepo: logRepo}
}

func (s *service) GerarEspelhoPonto(userID uint, empresaID uint, inicio, fim time.Time) (*EspelhoPonto, error) {
	if fim.Before(inicio) {
		inicio, fim = fim, inicio
	}

	usr, err := s.usuarioRead.FindByID(userID, empresaID)
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
		saldoDia := totalMinutos - cargaDia
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
