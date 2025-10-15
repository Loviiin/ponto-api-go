package bancohoras

// DashboardResponse é a estrutura completa de dados para a tela de banco de horas.
// swagger:model DashboardResponse
type DashboardResponse struct {
	// Saldo total acumulado do banco de horas em minutos
	// example: 120
	SaldoTotalMinutos int `json:"saldo_total_minutos"`
	// Histórico de alterações do banco de horas, ordenado por data desc
	Historico []HistoricoDia `json:"historico"`
}

// HistoricoDia representa um único lançamento no histórico do banco de horas.
// swagger:model HistoricoDia
type HistoricoDia struct {
	// Data do lançamento (formato: 2006-01-02)
	// example: 2025-10-01
	Data string `json:"data"`
	// É o saldo do dia (variação aplicada no dia)
	// example: 30
	ValorAlteradoMinutos int `json:"valor_alterado_minutos"`
	// É o saldo acumulado após este dia
	// example: 90
	SaldoResultanteMinutos int `json:"saldo_resultante_minutos"`
	// Motivo do lançamento
	// example: Fechamento automático do dia 2025-10-01
	Motivo string `json:"motivo"`
}
