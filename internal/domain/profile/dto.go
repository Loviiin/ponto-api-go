package profile

import "time"

// ProfileResponse representa os dados completos do perfil do usuário
type ProfileResponse struct {
	// Informações Pessoais
	ID        uint      `json:"id"`
	Nome      string    `json:"nome"`
	Email     string    `json:"email"`
	CPF       string    `json:"cpf"` // Será mascarado no service
	Telefone  string    `json:"telefone,omitempty"`
	Avatar    string    `json:"avatar,omitempty"`
	CreatedAt time.Time `json:"created_at"`

	// Informações de Trabalho
	Empresa  EmpresaInfo   `json:"empresa"`
	Cargo    *CargoInfo    `json:"cargo,omitempty"`
	Contrato *ContratoInfo `json:"contrato,omitempty"`

	// Permissões e Roles
	Permissoes []PermissaoInfo `json:"permissoes"`

	// Localidades
	Localidades []LocalidadeInfo `json:"localidades,omitempty"`
}

// EmpresaInfo contém informações básicas da empresa
type EmpresaInfo struct {
	ID          uint   `json:"id"`
	Nome        string `json:"nome"`
	CNPJ        string `json:"cnpj,omitempty"`
	RazaoSocial string `json:"razao_social,omitempty"`
}

// CargoInfo contém informações do cargo
type CargoInfo struct {
	ID        uint   `json:"id"`
	Nome      string `json:"nome"`
	Descricao string `json:"descricao,omitempty"`
}

// ContratoInfo contém informações do contrato
type ContratoInfo struct {
	ID                  uint       `json:"id"`
	DataInicio          time.Time  `json:"data_inicio"`
	DataFim             *time.Time `json:"data_fim,omitempty"`
	TipoContrato        string     `json:"tipo_contrato,omitempty"`
	CargaHorariaSemanal int        `json:"carga_horaria_semanal"`
	CargaHorariaMensal  int        `json:"carga_horaria_mensal"`
	HorarioEntrada      string     `json:"horario_entrada,omitempty"`
	HorarioSaida        string     `json:"horario_saida,omitempty"`
	IntervaloMinutos    int        `json:"intervalo_minutos"`
}

// PermissaoInfo contém informações de uma permissão
type PermissaoInfo struct {
	ID        uint   `json:"id"`
	Nome      string `json:"nome"`
	Descricao string `json:"descricao,omitempty"`
}

// LocalidadeInfo contém informações de uma localidade
type LocalidadeInfo struct {
	ID        uint    `json:"id"`
	Nome      string  `json:"nome"`
	Endereco  string  `json:"endereco"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Raio      int     `json:"raio"`
}

// UpdateProfileRequest representa os dados que podem ser atualizados pelo usuário
type UpdateProfileRequest struct {
	Nome     *string `json:"nome" binding:"omitempty,min=3,max=255"`
	Telefone *string `json:"telefone" binding:"omitempty,min=10,max=20"`
}

// ChangePasswordRequest representa a requisição de alteração de senha
type ChangePasswordRequest struct {
	SenhaAtual     string `json:"senha_atual" binding:"required,min=6"`
	NovaSenha      string `json:"nova_senha" binding:"required,min=6"`
	ConfirmarSenha string `json:"confirmar_senha" binding:"required,min=6"`
}

// StatsResponse representa as estatísticas de ponto do usuário
type StatsResponse struct {
	// Banco de Horas
	BancoHoras BancoHorasInfo `json:"banco_horas"`

	// Estatísticas do Mês Atual
	MesAtual MesAtualStats `json:"mes_atual"`

	// Estatísticas Gerais
	Geral GeralStats `json:"geral"`

	// Últimas Atividades
	UltimosRegistros []RegistroPontoInfo `json:"ultimos_registros"`
}

// BancoHorasInfo contém informações do banco de horas
type BancoHorasInfo struct {
	SaldoAtual        int       `json:"saldo_atual"`   // Em minutos
	SaldoFormatado    string    `json:"saldo_formato"` // Ex: "+08:30" ou "-02:15"
	UltimaAtualizacao time.Time `json:"ultima_atualizacao"`
}

// MesAtualStats contém estatísticas do mês atual
type MesAtualStats struct {
	Mes                   int     `json:"mes"`
	Ano                   int     `json:"ano"`
	DiasTrabalados        int     `json:"dias_trabalhados"`
	DiasAusentes          int     `json:"dias_ausentes"`
	DiasFaltantes         int     `json:"dias_faltantes"`    // Dias úteis restantes
	HorasTrabalhadas      int     `json:"horas_trabalhadas"` // Em minutos
	MediaHorasDia         int     `json:"media_horas_dia"`   // Em minutos
	TaxaPontualidade      float64 `json:"taxa_pontualidade"` // Percentual 0-100
	TotalAtrasos          int     `json:"total_atrasos"`
	TotalSaidasAdiantadas int     `json:"total_saidas_adiantadas"`
}

// GeralStats contém estatísticas gerais
type GeralStats struct {
	TotalPontosRegistrados  int     `json:"total_pontos_registrados"`
	TotalJustificativas     int     `json:"total_justificativas"`
	JustificativasPendentes int     `json:"justificativas_pendentes"`
	TaxaPontualizacaoGeral  float64 `json:"taxa_pontualizacao_geral"`
}

// RegistroPontoInfo contém informações de um registro de ponto
type RegistroPontoInfo struct {
	ID          uint       `json:"id"`
	Data        time.Time  `json:"data"`
	HoraEntrada *time.Time `json:"hora_entrada,omitempty"`
	HoraSaida   *time.Time `json:"hora_saida,omitempty"`
	Localidade  string     `json:"localidade,omitempty"`
	Observacao  string     `json:"observacao,omitempty"`
	Tipo        string     `json:"tipo"` // "normal", "atrasado", "falta", etc
}

// RecentActivityResponse representa atividades recentes do usuário
type RecentActivityResponse struct {
	Atividades []AtividadeInfo `json:"atividades"`
	Total      int             `json:"total"`
}

// AtividadeInfo representa uma atividade no sistema
type AtividadeInfo struct {
	ID        uint                   `json:"id"`
	Tipo      string                 `json:"tipo"` // "ponto", "justificativa", "perfil_atualizado", etc
	Descricao string                 `json:"descricao"`
	Data      time.Time              `json:"data"`
	Icone     string                 `json:"icone,omitempty"` // Para o frontend usar
	Detalhes  map[string]interface{} `json:"detalhes,omitempty"`
}

// CalendarioResponse representa o calendário de presença do usuário
type CalendarioResponse struct {
	Mes  int             `json:"mes"`
	Ano  int             `json:"ano"`
	Dias []CalendarioDia `json:"dias"`
}

// CalendarioDia representa um dia no calendário
type CalendarioDia struct {
	Dia              int    `json:"dia"`
	Status           string `json:"status"`            // "trabalhado", "ausente", "feriado", "fim_de_semana", "futuro"
	HorasTrabalhadas int    `json:"horas_trabalhadas"` // Em minutos
	Tipo             string `json:"tipo,omitempty"`    // "normal", "atrasado", "falta_justificada", etc
}

// PermissoesResponse representa as permissões do usuário
type PermissoesResponse struct {
	Permissoes []PermissaoDetalhada `json:"permissoes"`
	Total      int                  `json:"total"`
}

// PermissaoDetalhada representa uma permissão com todos os detalhes
type PermissaoDetalhada struct {
	ID          uint   `json:"id"`
	Nome        string `json:"nome"`
	Descricao   string `json:"descricao"`
	Categoria   string `json:"categoria"` // "ponto", "usuario", "relatorio", etc
	PodeVer     bool   `json:"pode_ver"`
	PodeCriar   bool   `json:"pode_criar"`
	PodeEditar  bool   `json:"pode_editar"`
	PodeDeletar bool   `json:"pode_deletar"`
}

// DocumentosResponse representa documentos disponíveis para download
type DocumentosResponse struct {
	Documentos []DocumentoInfo `json:"documentos"`
	Total      int             `json:"total"`
}

// DocumentoInfo representa um documento
type DocumentoInfo struct {
	ID          uint      `json:"id"`
	Tipo        string    `json:"tipo"` // "comprovante_ponto", "relatorio_banco_horas", etc
	Nome        string    `json:"nome"`
	Descricao   string    `json:"descricao,omitempty"`
	DataGeracao time.Time `json:"data_geracao"`
	URL         string    `json:"url"`     // URL para download
	Formato     string    `json:"formato"` // "pdf", "xlsx", etc
}

// PreferenciasResponse representa as preferências do usuário
type PreferenciasResponse struct {
	Tema         string            `json:"tema"`   // "light", "dark", "auto"
	Idioma       string            `json:"idioma"` // "pt-BR", "en-US", etc
	Notificacoes NotificacoesPrefs `json:"notificacoes"`
	FormatoHora  string            `json:"formato_hora"` // "24h", "12h"
	FormatoData  string            `json:"formato_data"` // "DD/MM/YYYY", "MM/DD/YYYY", etc
}

// NotificacoesPrefs representa preferências de notificações
type NotificacoesPrefs struct {
	Email               bool `json:"email"`
	Push                bool `json:"push"`
	JustificativaStatus bool `json:"justificativa_status"`
	LembretesPonto      bool `json:"lembretes_ponto"`
	AvisosBancoHoras    bool `json:"avisos_banco_horas"`
}

// UpdatePreferenciasRequest representa a requisição para atualizar preferências
type UpdatePreferenciasRequest struct {
	Tema         *string            `json:"tema" binding:"omitempty,oneof=light dark auto"`
	Idioma       *string            `json:"idioma" binding:"omitempty,oneof=pt-BR en-US"`
	Notificacoes *NotificacoesPrefs `json:"notificacoes,omitempty"`
	FormatoHora  *string            `json:"formato_hora" binding:"omitempty,oneof=24h 12h"`
	FormatoData  *string            `json:"formato_data" binding:"omitempty"`
}

// HistoricoLoginResponse representa o histórico de logins
type HistoricoLoginResponse struct {
	Logins []LoginInfo `json:"logins"`
	Total  int         `json:"total"`
}

// LoginInfo representa informações de um login
type LoginInfo struct {
	ID          uint      `json:"id"`
	Data        time.Time `json:"data"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	Dispositivo string    `json:"dispositivo"` // "desktop", "mobile", "tablet"
	Navegador   string    `json:"navegador"`
	Localizacao string    `json:"localizacao,omitempty"` // Cidade/País aproximado
	Sucesso     bool      `json:"sucesso"`
}
