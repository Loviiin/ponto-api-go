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

func Test_GetAddressByCEP_InvalidFormat(t *testing.T) {
	fb := &fakeBrasil{}
	fv := &fakeVia{}
	svc := NewService(fb, fv)

	_, err := svc.GetAddressByCEP("123") // not 8 digits
	if err == nil {
		t.Fatalf("expected error for invalid CEP format, got nil")
	}
	if err.Error() != "formato de CEP inválido" {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.called || fv.called {
		t.Fatalf("providers should not be called for invalid CEP: brasil=%v via=%v", fb.called, fv.called)
	}
}

func Test_GetAddressByCEP_BrasilAPISuccess(t *testing.T) {
	fb := &fakeBrasil{resp: &model.Localidade{CEP: "01001000", Cidade: "São Paulo", Estado: "SP"}}
	fv := &fakeVia{}
	svc := NewService(fb, fv)

	got, err := svc.GetAddressByCEP("01001-000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fb.called {
		t.Fatalf("expected BrasilAPI to be called")
	}
	if fv.called {
		t.Fatalf("did not expect ViaCEP to be called when BrasilAPI succeeds")
	}
	if fb.lastCEP != "01001000" {
		t.Fatalf("expected normalized CEP passed to BrasilAPI, got %s", fb.lastCEP)
	}
	if got.CEP != "01001000" || got.Cidade != "São Paulo" || got.Estado != "SP" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func Test_GetAddressByCEP_FallbackToViaCEP(t *testing.T) {
	fb := &fakeBrasil{err: errors.New("service down")}
	fv := &fakeVia{resp: &model.Localidade{CEP: "01001000", Cidade: "São Paulo", Estado: "SP"}}
	svc := NewService(fb, fv)

	got, err := svc.GetAddressByCEP("01001-000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fb.called || !fv.called {
		t.Fatalf("expected both providers to be called: brasil=%v via=%v", fb.called, fv.called)
	}
	if fv.lastCEP != "01001000" {
		t.Fatalf("expected normalized CEP passed to ViaCEP, got %s", fv.lastCEP)
	}
	if got.Cidade != "São Paulo" || got.Estado != "SP" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func Test_GetAddressByCEP_BothFail(t *testing.T) {
	fb := &fakeBrasil{err: errors.New("not found")}
	fv := &fakeVia{err: errors.New("not found")}
	svc := NewService(fb, fv)

	_, err := svc.GetAddressByCEP("01001-000")
	if err == nil {
		t.Fatalf("expected error when both providers fail")
	}
}
