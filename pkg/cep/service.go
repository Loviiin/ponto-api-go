package cep

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/brasilapi"
	"github.com/Loviiin/ponto-api-go/pkg/viacep"
)

// Service é a interface para o serviço de consulta de CEP com fallback
type Service interface {
	GetCEPInfo(cep string) (*model.Localidade, error)
	// Alias para compatibilidade com especificações anteriores
	// Retorna o mesmo resultado de GetCEPInfo
	GetAddressByCEP(cep string) (*model.Localidade, error)
}

type service struct {
	brasilAPIClient brasilapi.Client
	viaCEPClient    viacep.Client
}

// NewService cria uma nova instância do serviço de CEP
// que utiliza ViaCEP como primário e BrasilAPI como fallback
func NewService(brasilAPIClient brasilapi.Client, viaCEPClient viacep.Client) Service {
	return &service{
		brasilAPIClient: brasilAPIClient,
		viaCEPClient:    viaCEPClient,
	}
}

var nonDigitsRegex = regexp.MustCompile(`\D`)

// ErrInvalidCEP é retornado quando o CEP não possui 8 dígitos válidos.
var ErrInvalidCEP = errors.New("formato de CEP inválido")

// normalizeCEP remove caracteres não numéricos e valida se há 8 dígitos.
func normalizeCEP(cep string) (string, error) {
	onlyDigits := nonDigitsRegex.ReplaceAllString(cep, "")
	if len(onlyDigits) != 8 {
		return "", ErrInvalidCEP
	}
	return onlyDigits, nil
}

// GetCEPInfo tenta buscar informações do CEP primeiro no ViaCEP
// e, em caso de falha, usa a BrasilAPI como fallback
func (s *service) GetCEPInfo(cep string) (*model.Localidade, error) {
	// Normaliza e valida o CEP antes de chamar provedores externos
	normalized, err := normalizeCEP(cep)
	if err != nil {
		return nil, err
	}
	// Tenta primeiro com ViaCEP (mais rápido e confiável)
	localidade, err := s.viaCEPClient.GetCEPInfo(normalized)
	if err == nil && localidade != nil {
		log.Printf("[CEP] Consulta %s: ViaCEP (primário) SUCESSO", normalized)
		return localidade, nil
	}

	// Se ViaCEP falhou, armazena o erro para log
	viaCEPError := err
	log.Printf("[CEP] Consulta %s: ViaCEP falhou (%v), tentando fallback BrasilAPI...", normalized, viaCEPError)

	// Fallback: tenta com BrasilAPI
	localidade, err = s.brasilAPIClient.GetCEPInfo(normalized)
	if err != nil {
		log.Printf("[CEP] Consulta %s: BrasilAPI (fallback) também falhou (%v)", normalized, err)
		// Ambas as APIs falharam
		return nil, fmt.Errorf("falha ao buscar CEP. ViaCEP: %v | BrasilAPI: %w", viaCEPError, err)
	}

	log.Printf("[CEP] Consulta %s: BrasilAPI (fallback) SUCESSO", normalized)
	// BrasilAPI retornou com sucesso
	return localidade, nil
}

// GetAddressByCEP é um alias para GetCEPInfo para manter compatibilidade
func (s *service) GetAddressByCEP(cep string) (*model.Localidade, error) {
	return s.GetCEPInfo(cep)
}
