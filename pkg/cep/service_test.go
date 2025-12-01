package cep

import (
	"errors"
	"testing"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/brasilapi"
	"github.com/Loviiin/ponto-api-go/pkg/viacep"
)

// --- Fakes ---
type fakeBrasil struct {
	called  bool
	lastCEP string
	resp    *model.Localidade
	err     error
}

func (f *fakeBrasil) GetCEPInfo(cep string) (*model.Localidade, error) {
	f.called = true
	f.lastCEP = cep
	return f.resp, f.err
}

var _ brasilapi.Client = (*fakeBrasil)(nil)

type fakeVia struct {
	called  bool
	lastCEP string
	resp    *model.Localidade
	err     error
}

func (f *fakeVia) GetCEPInfo(cep string) (*model.Localidade, error) {
	f.called = true
	f.lastCEP = cep
	return f.resp, f.err
}

var _ viacep.Client = (*fakeVia)(nil)

func Test_ObterEnderecoPorCEP_FormatoInvalido(t *testing.T) {
	fb := &fakeBrasil{}
	fv := &fakeVia{}
	svc := NewService(fb, fv)

	_, err := svc.GetAddressByCEP("123") // não tem 8 dígitos
	if err == nil {
		t.Fatalf("esperava erro para formato de CEP inválido, recebeu nil")
	}
	if err.Error() != "formato de CEP inválido" {
		t.Fatalf("erro inesperado: %v", err)
	}
	if fb.called || fv.called {
		t.Fatalf("provedores não deveriam ser chamados para CEP inválido: brasil=%v via=%v", fb.called, fv.called)
	}
}

func Test_ObterEnderecoPorCEP_ViaCEPSucesso(t *testing.T) {
	// Configuração: ViaCEP retorna sucesso, BrasilAPI não deve ser chamado
	fb := &fakeBrasil{}
	fv := &fakeVia{resp: &model.Localidade{CEP: "01001000", Cidade: "São Paulo", Estado: "SP"}}
	svc := NewService(fb, fv)

	got, err := svc.GetAddressByCEP("01001-000")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !fv.called {
		t.Fatalf("esperava que ViaCEP fosse chamado")
	}
	if fb.called {
		t.Fatalf("não esperava que BrasilAPI fosse chamado quando ViaCEP tem sucesso")
	}
	if fv.lastCEP != "01001000" {
		t.Fatalf("esperava CEP normalizado passado para ViaCEP, recebeu %s", fv.lastCEP)
	}
	if got.CEP != "01001000" || got.Cidade != "São Paulo" || got.Estado != "SP" {
		t.Fatalf("resultado inesperado: %+v", got)
	}
}

func Test_ObterEnderecoPorCEP_FallbackParaBrasilAPI(t *testing.T) {
	// Configuração: ViaCEP falha, deve tentar BrasilAPI
	fb := &fakeBrasil{resp: &model.Localidade{CEP: "01001000", Cidade: "São Paulo", Estado: "SP"}}
	fv := &fakeVia{err: errors.New("serviço fora do ar")}
	svc := NewService(fb, fv)

	got, err := svc.GetAddressByCEP("01001-000")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !fv.called || !fb.called {
		t.Fatalf("esperava que ambos os provedores fossem chamados: via=%v brasil=%v", fv.called, fb.called)
	}
	if fb.lastCEP != "01001000" {
		t.Fatalf("esperava CEP normalizado passado para BrasilAPI, recebeu %s", fb.lastCEP)
	}
	if got.Cidade != "São Paulo" || got.Estado != "SP" {
		t.Fatalf("resultado inesperado: %+v", got)
	}
}

func Test_ObterEnderecoPorCEP_AmbosFalham(t *testing.T) {
	fb := &fakeBrasil{err: errors.New("não encontrado")}
	fv := &fakeVia{err: errors.New("não encontrado")}
	svc := NewService(fb, fv)

	_, err := svc.GetAddressByCEP("01001-000")
	if err == nil {
		t.Fatalf("esperava erro quando ambos os provedores falham")
	}
}
