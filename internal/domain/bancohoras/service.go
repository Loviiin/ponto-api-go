package bancohoras

import (
	"sort"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
)

type BancoHorasService interface {
	CalcularSaldoParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (int, error)
	FecharDiaParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (*model.Usuario, error)
}

type bancoHorasService struct {
	pontoRepo   ponto.RegistroPontoRepository
	usuarioRepo usuario.UsuarioRepository
}

func NewBancoHorasService(pontoRepo ponto.RegistroPontoRepository, userRepo usuario.UsuarioRepository) BancoHorasService {
	return &bancoHorasService{
		pontoRepo:   pontoRepo,
		usuarioRepo: userRepo,
	}
}

func (s *bancoHorasService) CalcularSaldoParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (int, error) {
	user, err := s.usuarioRepo.FindByID(usuarioID, empresaID)
	if err != nil {
		return 0, err
	}
	pontos, err := s.pontoRepo.FindPontosByUserIDAndDate(user.ID, dia)
	if err != nil {
		return 0, err
	}
	doDia, err := CalcularSaldoDoDia(pontos, user.Cargo)
	if err != nil {
		return 0, err
	}

	return doDia, err
}

func CalcularSaldoDoDia(pontosDoDia []model.RegistroPonto, cargoDoUsuario model.Cargo) (saldoEmMinutos int, err error) {
	sort.Slice(pontosDoDia, func(i, j int) bool {
		return pontosDoDia[i].Timestamp.Before(pontosDoDia[j].Timestamp)
	})

	var totalTrabalhadoEmMinutos float64 = 0
	for i := 0; i < len(pontosDoDia)-1; i += 2 {
		entrada := pontosDoDia[i].Timestamp
		saida := pontosDoDia[i+1].Timestamp

		duracao := saida.Sub(entrada).Minutes()
		totalTrabalhadoEmMinutos += duracao
	}

	saldo := totalTrabalhadoEmMinutos - float64(cargoDoUsuario.CargaHorariaDiariaMinutos)

	return int(saldo), nil
}

func (s *bancoHorasService) FecharDiaParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (*model.Usuario, error) {
	saldoDoDia, err := s.CalcularSaldoParaUsuario(usuarioID, empresaID, dia)
	if err != nil {
		return nil, err
	}

	usuarioAtual, err := s.usuarioRepo.FindByID(usuarioID, empresaID)
	if err != nil {
		return nil, err
	}

	novoSaldoTotal := usuarioAtual.SaldoBancoHorasMinutos + saldoDoDia

	dadosParaAtualizar := map[string]interface{}{
		"saldo_banco_horas_minutos": novoSaldoTotal,
	}

	err = s.usuarioRepo.Update(usuarioID, empresaID, dadosParaAtualizar)
	if err != nil {
		return nil, err
	}

	usuarioAtual.SaldoBancoHorasMinutos = novoSaldoTotal
	return usuarioAtual, nil
}
