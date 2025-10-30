package justificativa

import (
	"errors"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type Service interface {
	SolicitarAjuste(solicitacao *model.Justificativa) error
	SolicitarCorrecaoPonto(pontoID uint, novaDataHora time.Time, descricao string, usuarioID, empresaID uint) error
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
	// Validação: PONTO_FALTANTE não deve ter ponto_id
	if solicitacao.Tipo == "PONTO_FALTANTE" && solicitacao.PontoID != nil {
		return errors.New("justificativa do tipo PONTO_FALTANTE não pode ter ponto_id")
	}

	// Validação: CORRECAO_PONTO deve ter ponto_id
	if solicitacao.Tipo == "CORRECAO_PONTO" && solicitacao.PontoID == nil {
		return errors.New("justificativa do tipo CORRECAO_PONTO deve ter ponto_id")
	}

	solicitacao.Status = "PENDENTE"
	return s.justificativaRepo.Create(solicitacao)
}

// SolicitarCorrecaoPonto cria uma justificativa específica para correção de ponto existente
func (s *service) SolicitarCorrecaoPonto(pontoID uint, novaDataHora time.Time, descricao string, usuarioID, empresaID uint) error {
	// Validar se o ponto existe e pertence ao usuário
	ponto, err := s.pontoRepo.FindPontoByID(pontoID, empresaID)
	if err != nil {
		return errors.New("ponto não encontrado")
	}

	if ponto.UsuarioID != usuarioID {
		return errors.New("você não tem permissão para corrigir este ponto")
	}

	// Criar justificativa do tipo CORRECAO_PONTO
	justificativa := &model.Justificativa{
		Tipo:           "CORRECAO_PONTO",
		PontoID:        &pontoID,
		DataOcorrencia: novaDataHora,
		Descricao:      descricao,
		UsuarioID:      usuarioID,
		EmpresaID:      empresaID,
		Status:         "PENDENTE",
	}

	return s.justificativaRepo.Create(justificativa)
}

func (s *service) ListarPendentes(empresaID uint) ([]model.Justificativa, error) {
	return s.justificativaRepo.FindByStatus(empresaID, "PENDENTE")
}

func (s *service) ListarPorUsuario(usuarioID uint, empresaID uint) ([]model.Justificativa, error) {
	return s.justificativaRepo.FindByUsuarioID(usuarioID, empresaID)
}

func (s *service) AprovarReprovar(justificativaID, empresaID, aprovadorID uint, aprovado bool, motivoReprovacao string) (*model.Justificativa, error) {
	// Usamos uma transação para garantir que a atualização da justificativa e a criação/atualização do ponto
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

			pontoRepoTx := s.pontoRepo.WithTransaction(tx)

			// LÓGICA DIFERENCIADA POR TIPO DE JUSTIFICATIVA
			if justificativa.Tipo == "CORRECAO_PONTO" && justificativa.PontoID != nil {
				// Tipo 1: CORREÇÃO DE PONTO EXISTENTE
				// Buscar o ponto que será corrigido
				pontoExistente, err := pontoRepoTx.FindPontoByID(*justificativa.PontoID, empresaID)
				if err != nil {
					return errors.New("ponto a ser corrigido não encontrado")
				}

				// Usar NovoHorario se disponível, senão DataOcorrencia
				novoTimestamp := justificativa.DataOcorrencia
				if justificativa.NovoHorario != nil {
					novoTimestamp = *justificativa.NovoHorario
				}

				// Atualizar o timestamp do ponto existente
				pontoExistente.Timestamp = novoTimestamp
				pontoExistente.Metodo = "AJUSTE_APROVADO"
				pontoExistente.Status = "APROVADO"
				pontoExistente.JustificativaID = &justificativa.ID

				if err := pontoRepoTx.UpdatePonto(pontoExistente); err != nil {
					return errors.New("erro ao atualizar ponto: " + err.Error())
				}

			} else {
				// Tipo 2: PONTO FALTANTE (comportamento original)
				// Criar um novo ponto com status APROVADO
				timestampNovo := justificativa.DataOcorrencia
				if justificativa.NovoHorario != nil {
					timestampNovo = *justificativa.NovoHorario
				}

				pontoRegistrado := &model.RegistroPonto{
					UsuarioID:       justificativa.UsuarioID,
					EmpresaID:       empresaID,
					Timestamp:       timestampNovo,
					Metodo:          "AJUSTE_APROVADO",
					Localizacao:     "N/A",
					Status:          "APROVADO",
					JustificativaID: &justificativa.ID,
				}
				if err := pontoRepoTx.SavePonto(pontoRegistrado); err != nil {
					return err
				}
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
