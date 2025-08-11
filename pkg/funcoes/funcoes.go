package funcoes

import (
	"strconv"
)

type FuncoesInterface interface {
	StrParaUint(id string) (uint, error)
}

type funcoes struct{}

func NewFuncoes() FuncoesInterface {
	return &funcoes{}
}

func (f *funcoes) StrParaUint(id string) (uint, error) {
	idUint64, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(idUint64), nil
}
