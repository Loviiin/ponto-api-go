package funcoes

import (
	"errors"
	"github.com/gin-gonic/gin"
	"strconv"
)

type FuncoesInterface interface {
	StrParaUint(id string) (uint, error)
	GetUintIDFromContext(c *gin.Context, key string) (uint, error)
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

func (f *funcoes) GetUintIDFromContext(c *gin.Context, key string) (uint, error) {
	valorID, existe := c.Get(key)
	if !existe {
		return 0, errors.New("ID com a chave '" + key + "' não encontrado no contexto")
	}

	idString, ok := valorID.(string)
	if !ok {
		return 0, errors.New("ID com a chave '" + key + "' no contexto não é uma string")
	}

	id, err := f.StrParaUint(idString)
	if err != nil {
		return 0, errors.New("ID com a chave '" + key + "' no contexto não é um número válido")
	}

	return id, nil
}
