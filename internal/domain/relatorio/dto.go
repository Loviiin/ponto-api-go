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

// --- DTOs para Relatório Geral ---

// RelatorioDiaDTO representa os dados de um dia específico no relatório
type RelatorioDiaDTO struct {
	Data                   string            `json:"data"`
	Marcacoes              []ParMarcacao     `json:"marcacoes"`
	TotalTrabalhadoMinutos int               `json:"total_trabalhado_minutos"`
	TotalTrabalhadoHHMM    string            `json:"total_trabalhado_hhmm"`
	CargaEsperadaMinutos   int               `json:"carga_esperada_minutos"`
	CargaEsperadaHHMM      string            `json:"carga_esperada_hhmm"`
	HorasExtrasMinutos     int               `json:"horas_extras_minutos"`    // positivo se trabalhou mais
	HorasFaltantesMinutos  int               `json:"horas_faltantes_minutos"` // positivo se trabalhou menos
	SaldoBancoHorasMinutos int               `json:"saldo_banco_horas_minutos"`
	Justificativa          *JustificativaDTO `json:"justificativa,omitempty"`
	Inconsistente          bool              `json:"inconsistente"`
}

// JustificativaDTO representa uma justificativa no relatório
type JustificativaDTO struct {
	ID        uint   `json:"id"`
	Tipo      string `json:"tipo"`
	Descricao string `json:"descricao"`
	Status    string `json:"status"`
}

// ResumoUsuarioDTO contém os totais e resumo do período para um usuário
type ResumoUsuarioDTO struct {
	TotalHorasTrabalhadasMinutos  int    `json:"total_horas_trabalhadas_minutos"`
	TotalHorasTrabalhadasHHMM     string `json:"total_horas_trabalhadas_hhmm"`
	TotalHorasExtrasMinutos       int    `json:"total_horas_extras_minutos"`
	TotalHorasExtrasHHMM          string `json:"total_horas_extras_hhmm"`
	TotalHorasFaltantesMinutos    int    `json:"total_horas_faltantes_minutos"`
	TotalHorasFaltantesHHMM       string `json:"total_horas_faltantes_hhmm"`
	SaldoBancoHorasInicialMinutos int    `json:"saldo_banco_horas_inicial_minutos"`
	SaldoBancoHorasFinalMinutos   int    `json:"saldo_banco_horas_final_minutos"`
}

// UsuarioInfoDTO contém informações básicas do usuário
type UsuarioInfoDTO struct {
	ID        uint   `json:"id"`
	Nome      string `json:"nome"`
	CPF       string `json:"cpf"`
	Email     string `json:"email"`
	CargoNome string `json:"cargo_nome,omitempty"`
}

// RelatorioUsuarioDTO representa o relatório completo de um usuário
type RelatorioUsuarioDTO struct {
	Usuario UsuarioInfoDTO    `json:"usuario"`
	Periodo PeriodoInterval   `json:"periodo"`
	Dias    []RelatorioDiaDTO `json:"dias"`
	Resumo  ResumoUsuarioDTO  `json:"resumo"`
}

// RelatorioGeralDTO representa o relatório geral que pode conter um ou mais usuários
type RelatorioGeralDTO struct {
	Periodo  PeriodoInterval       `json:"periodo"`
	Usuarios []RelatorioUsuarioDTO `json:"usuarios"`
}
