package justificativa

import (
	"errors"

	"github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type Service interface {
	SolicitarAjuste(solicitacao *model.Justificativa) error
	ListarPendentes(empresaID uint) ([]model.Justificativa, error)
	ListarPorUsuario(usuarioID uint, empresaID uint) ([]model.Justificativa, error)
	AprovarReprovar(justificativaID, empresaID, aprovadorID uint, aprovado bool, motivoReprovacao string) (*model.Justificativa, error)
	CancelarSolicitacao(justificativaID, empresaID, usuarioID uint) error
}

type service struct {
	justificativaRepo Repository
	// MUDANÇA: Trocamos o serviço pelo repositório de ponto
	pontoRepo ponto.RegistroPontoRepository
	db        *gorm.DB
}

func NewService(repo Repository, pontoRepo ponto.RegistroPontoRepository, db *gorm.DB) Service {
	return &service{
		justificativaRepo: repo,
		pontoRepo:         pontoRepo, // Atribuímos o repositório
		db:                db,
	}
}

func (s *service) SolicitarAjuste(solicitacao *model.Justificativa) error {
	solicitacao.Status = "PENDENTE"
	return s.justificativaRepo.Create(solicitacao)
}

func (s *service) ListarPendentes(empresaID uint) ([]model.Justificativa, error) {
	return s.justificativaRepo.FindByStatus(empresaID, "PENDENTE")
}

func (s *service) ListarPorUsuario(usuarioID uint, empresaID uint) ([]model.Justificativa, error) {
	return s.justificativaRepo.FindByUsuarioID(usuarioID, empresaID)
}

func (s *service) AprovarReprovar(justificativaID, empresaID, aprovadorID uint, aprovado bool, motivoReprovacao string) (*model.Justificativa, error) {
	// Usamos uma transação para garantir que a atualização da justificativa e a criação do ponto
	var justificativaProcessada *model.Justificativa

	err := s.db.Transaction(func(tx *gorm.DB) error {
		repoTx := s.justificativaRepo.WithTransaction(tx)

		justificativa, err := repoTx.FindByID(justificativaID, empresaID)
		if err != nil {
			return errors.New("justificativa não encontrada")
		}

		if justificativa.Status != "PENDENTE" {
			return errors.New("esta solicitação já foi processada")
		}

		justificativa.AprovadorID = &aprovadorID

		if aprovado {
			justificativa.Status = "APROVADO"

			// Lógica de criação do ponto movida para cá
			pontoRepoTx := s.pontoRepo.WithTransaction(tx)
			pontoRegistrado := &model.RegistroPonto{
				UsuarioID:       justificativa.UsuarioID,
				EmpresaID:       empresaID,
				Timestamp:       justificativa.DataOcorrencia,
				Metodo:          "AJUSTE_APROVADO",
				Localizacao:     "N/A",
				JustificativaID: &justificativa.ID,
			}
			if err := pontoRepoTx.SavePonto(pontoRegistrado); err != nil {
				return err
			}

		} else {
			justificativa.Status = "REPROVADO"
			if motivoReprovacao != "" {
				justificativa.ObservacaoAprovador = motivoReprovacao
			}
		}

		if err := repoTx.Update(justificativa); err != nil {
			return err
		}

		justificativaProcessada = justificativa
		return nil
	})

	return justificativaProcessada, err
}

// CancelarSolicitacao permite que o próprio usuário cancele sua solicitação pendente
func (s *service) CancelarSolicitacao(justificativaID, empresaID, usuarioID uint) error {
	justificativa, err := s.justificativaRepo.FindByID(justificativaID, empresaID)
	if err != nil {
		return errors.New("justificativa não encontrada")
	}

	// Validar que é o próprio usuário tentando cancelar
	if justificativa.UsuarioID != usuarioID {
		return errors.New("você só pode cancelar suas próprias solicitações")
	}

	// Validar que está pendente
	if justificativa.Status != "PENDENTE" {
		return errors.New("apenas solicitações pendentes podem ser canceladas")
	}

	// Atualizar status para CANCELADO
	justificativa.Status = "CANCELADO"
	justificativa.ObservacaoAprovador = "Cancelado pelo próprio usuário"

	if err := s.justificativaRepo.Update(justificativa); err != nil {
		return errors.New("erro ao cancelar solicitação")
	}

	return nil
}
