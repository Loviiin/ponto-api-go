package scheduler

import (
	"log/slog"
	"os"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/bancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	bancoHorasService bancohoras.BancoHorasService
	usuarioService    usuario.UsuarioService
	logger            *slog.Logger
}

func NewScheduler(bancohorasService bancohoras.BancoHorasService, usuarioService usuario.UsuarioService, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		bancoHorasService: bancohorasService,
		usuarioService:    usuarioService,
		logger:            logger,
	}
}

func (s *Scheduler) Start() {
	// Carrega o timezone do Brasil para o scheduler
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		s.logger.Warn("AVISO: Não foi possível carregar timezone America/Sao_Paulo, usando Local", slog.Any("error", err))
		loc = time.Local
	}

	// Cria o cron com timezone Brasil
	c := cron.New(cron.WithLocation(loc))

	_, err = c.AddFunc("0 1 * * *", s.executarFechamentoDiario)
	if err != nil {
		// Fatal error during startup: scheduler is critical for banco de horas
		// This happens only during app initialization, so os.Exit is appropriate
		s.logger.Error("Erro ao agendar a tarefa de fechamento diário", slog.Any("error", err))
		os.Exit(1)
	}

	c.Start()

	s.logger.Info("Agendador de tarefas iniciado", slog.String("timezone", loc.String()), slog.String("schedule", "01:00 BRT"))
}

func (s *Scheduler) executarFechamentoDiario() {
	s.logger.Info("Iniciando tarefa agendada: Fechamento diário do banco de horas")

	// Usa o timezone Brasil para calcular o dia anterior
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local
	}

	agora := time.Now().In(loc)
	diaAnterior := agora.AddDate(0, 0, -1)

	usuarios, err := s.usuarioService.FindAll()
	if err != nil {
		s.logger.Error("SCHEDULER: Erro ao buscar usuários para o fechamento diário", slog.Any("error", err))
		return
	}

	s.logger.Info("Encontrados usuários para processar", slog.Int("count", len(usuarios)))

	for _, usr := range usuarios {
		_, err := s.bancoHorasService.FecharDiaParaUsuario(usr.ID, usr.Contrato.EmpresaID, diaAnterior)
		if err != nil {
			s.logger.Error("SCHEDULER: Erro ao fechar o dia para o usuário",
				slog.Uint64("usuario_id", uint64(usr.ID)),
				slog.Any("error", err))
		} else {
			s.logger.Info("SCHEDULER: Fechamento do dia concluído",
				slog.String("date", diaAnterior.Format("2006-01-02")),
				slog.Uint64("usuario_id", uint64(usr.ID)))
		}
	}

	s.logger.Info("Tarefa agendada: Fechamento diário concluído")
}
