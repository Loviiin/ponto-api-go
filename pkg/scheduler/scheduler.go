package scheduler

import (
	"github.com/Loviiin/ponto-api-go/internal/domain/bancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/robfig/cron/v3"
	"log"
	"time"
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
	c := cron.New()

	_, err := c.AddFunc("0 1 * * *", s.executarFechamentoDiario)
	if err != nil {
		log.Fatalf("Erro ao agendar a tarefa de fechamento diário: %v", err)
	}

	c.Start()

	log.Println("Agendador de tarefas iniciado. O fechamento diário será executado à 01:00.")
}

func (s *Scheduler) executarFechamentoDiario() {
	log.Println("Iniciando tarefa agendada: Fechamento diário do banco de horas...")

	diaAnterior := time.Now().AddDate(0, 0, -1)

	usuarios, err := s.usuarioService.FindAll()
	if err != nil {
		log.Printf("SCHEDULER: Erro ao buscar usuários para o fechamento diário: %v", err)
		return
	}

	log.Printf("Encontrados %d usuários para processar.", len(usuarios))

	for _, usr := range usuarios {
		_, err := s.bancoHorasService.FecharDiaParaUsuario(usr.ID, usr.EmpresaID, diaAnterior)
		if err != nil {
			log.Printf("SCHEDULER: Erro ao fechar o dia para o usuário ID %d: %v", usr.ID, err)
		} else {
			log.Printf("SCHEDULER: Fechamento do dia %s concluído para o usuário ID %d.", diaAnterior.Format("2006-01-02"), usr.ID)
		}
	}

	log.Println("Tarefa agendada: Fechamento diário concluído.")
}
