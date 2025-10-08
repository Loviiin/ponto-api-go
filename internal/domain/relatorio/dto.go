package relatorio

import "time"

type PeriodoInterval struct {
	Inicio string `json:"inicio"`
	Fim    string `json:"fim"`
}

type ParMarcacao struct {
	Entrada *time.Time `json:"entrada,omitempty"`
	Saida   *time.Time `json:"saida,omitempty"`
}

type DiaEspelho struct {
	Data                   string        `json:"data"`
	Marcacoes              []ParMarcacao `json:"marcacoes"`
	TotalTrabalhadoMinutos int           `json:"total_trabalhado_minutos"`
	TotalTrabalhadoHHMM    string        `json:"total_trabalhado_hhmm"`
	CargaPlanejadaMinutos  int           `json:"carga_planejada_minutos"`
	SaldoDiaMinutos        int           `json:"saldo_dia_minutos"`
	SaldoAcumuladoMinutos  int           `json:"saldo_acumulado_minutos"`
	Fechado                bool          `json:"fechado"`
	LogMotivo              string        `json:"log_motivo,omitempty"`
	Inconsistente          bool          `json:"inconsistente"`
}

type TotaisPeriodo struct {
	TrabalhadoMinutos int    `json:"trabalhado_minutos"`
	TrabalhadoHHMM    string `json:"trabalhado_hhmm"`
	SaldoFinalMinutos int    `json:"saldo_final_minutos"`
}

type EspelhoPonto struct {
	UsuarioID          uint            `json:"usuario_id"`
	Periodo            PeriodoInterval `json:"periodo"`
	Dias               []DiaEspelho    `json:"dias"`
	Totais             TotaisPeriodo   `json:"totais"`
	DiasInconsistentes []string        `json:"dias_inconsistentes"`
}

// GerarEspelhoResponse é o payload retornado pelos endpoints de espelho de ponto.
// @Description Estrutura consolidada do espelho de ponto para um intervalo de datas.
type GerarEspelhoResponse = EspelhoPonto
