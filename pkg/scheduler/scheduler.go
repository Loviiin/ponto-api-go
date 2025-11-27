package scheduler

import (
	"log"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/bancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	bancoHorasService bancohoras.BancoHorasService
	usuarioService    usuario.UsuarioService
}

func NewScheduler(bancohorasService bancohoras.BancoHorasService, usuarioService usuario.UsuarioService) *Scheduler {
	return &Scheduler{
		bancoHorasService: bancohorasService,
		usuarioService:    usuarioService,
	}
}

func (s *Scheduler) Start() {
	// Carrega o timezone do Brasil para o scheduler
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Printf("AVISO: Não foi possível carregar timezone America/Sao_Paulo, usando Local: %v", err)
		loc = time.Local
	}

	// Cria o cron com timezone Brasil
	c := cron.New(cron.WithLocation(loc))

	_, err = c.AddFunc("0 1 * * *", s.executarFechamentoDiario)
	if err != nil {
		log.Fatalf("Erro ao agendar a tarefa de fechamento diário: %v", err)
	}

	c.Start()

	log.Printf("Agendador de tarefas iniciado com timezone %s. O fechamento diário será executado à 01:00 BRT.", loc.String())
}

func (s *Scheduler) executarFechamentoDiario() {
	log.Println("Iniciando tarefa agendada: Fechamento diário de banco de horas.")

	// Calcula o dia anterior (ontem)
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	if loc == nil {
		loc = time.Local
	}
	diaAnterior := time.Now().In(loc).AddDate(0, 0, -1)

	// Busca todos os usuários ativos
	usuarios, err := s.usuarioService.FindAll()
	if err != nil {
		log.Printf("SCHEDULER: Erro ao buscar usuários ativos: %v", err)
		return
	}

	log.Printf("Encontrados %d usuários para processar.", len(usuarios))

	for _, usr := range usuarios {
		_, err := s.bancoHorasService.FecharDiaParaUsuario(usr.ID, usr.Contrato.EmpresaID, diaAnterior)
		if err != nil {
			log.Printf("SCHEDULER: Erro ao fechar o dia para o usuário ID %d: %v", usr.ID, err)
		} else {
			log.Printf("SCHEDULER: Fechamento do dia %s concluído para o usuário ID %d.", diaAnterior.Format("2006-01-02"), usr.ID)
		}
	}

	log.Println("Tarefa agendada: Fechamento diário concluído.")
}
