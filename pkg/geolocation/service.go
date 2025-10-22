package geolocation

import (
	"fmt"
	"log"
	"strings"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/brasilapi"
	"github.com/Loviiin/ponto-api-go/pkg/cep"
	"github.com/Loviiin/ponto-api-go/pkg/distancematrix"
)

type Service interface {
	GetLocationFromCEP(cep string) (*model.Localidade, error)
}

type service struct {
	cepService           cep.Service
	distanceMatrixClient distancematrix.Client
	brasilAPIClient      brasilapi.Client
}

func NewService(cepService cep.Service, distanceMatrixClient distancematrix.Client, brasilAPIClient brasilapi.Client) Service {
	return &service{
		cepService:           cepService,
		distanceMatrixClient: distanceMatrixClient,
		brasilAPIClient:      brasilAPIClient,
	}
}

func (s *service) GetLocationFromCEP(cep string) (*model.Localidade, error) {
	// ETAPA 1: Tentativa Primária - ViaCEP + Distance Matrix AI
	log.Printf("[Geo] CEP %s: Iniciando busca com cadeia primária (ViaCEP + Distance Matrix AI)", cep)

	// 1.1: Buscar endereço textual via cepService (ViaCEP primário, BrasilAPI fallback)
	localidadeEndereco, err := s.cepService.GetAddressByCEP(cep)
	if err != nil {
		// Se cepService falhou completamente, pular para erro final
		log.Printf("[Geo] CEP %s: Erro ao buscar endereço via cepService: %v", cep, err)
		return nil, fmt.Errorf("falha ao buscar informações do CEP: %w", err)
	}

	// 1.2: Construir string de busca otimizada para Distance Matrix AI
	// Remover strings vazias e evitar duplicatas
	var parts []string
	if localidadeEndereco.Logradouro != "" && localidadeEndereco.Logradouro != "Rua" {
		// Remover prefixo duplicado (ex: "Rua Rua X" -> "Rua X")
		logradouro := strings.TrimSpace(localidadeEndereco.Logradouro)
		if strings.HasPrefix(strings.ToLower(logradouro), "rua ") && strings.HasPrefix(strings.ToLower(strings.TrimPrefix(logradouro, "Rua ")), "rua ") {
			logradouro = strings.TrimPrefix(logradouro, "Rua ")
		}
		parts = append(parts, logradouro)
	}
	if localidadeEndereco.Bairro != "" {
		parts = append(parts, strings.TrimSpace(localidadeEndereco.Bairro))
	}
	if localidadeEndereco.Cidade != "" {
		parts = append(parts, strings.TrimSpace(localidadeEndereco.Cidade))
	}
	if localidadeEndereco.Estado != "" {
		parts = append(parts, strings.TrimSpace(localidadeEndereco.Estado))
	}
	addressString := strings.Join(parts, ", ") + ", Brasil"
	log.Printf("[Geo] CEP %s: Endereço obtido, buscando coordenadas via Distance Matrix AI: %s", cep, addressString)

	// 1.3: Chamar Distance Matrix AI para obter coordenadas
	geocodeResp, err := s.distanceMatrixClient.GeocodeAddress(addressString)
	if err != nil {
		// Falhou na Distance Matrix AI - tentar fallback
		log.Printf("[Geo] CEP %s: Falha ao obter coordenadas da Distance Matrix AI (%v), tentando fallback com BrasilAPI...", cep, err)
		goto TentarFallback
	}

	// 1.4: Verificar se Distance Matrix AI retornou resultados válidos
	if geocodeResp.Status != "OK" || len(geocodeResp.Result) == 0 {
		log.Printf("[Geo] CEP %s: Distance Matrix AI não retornou resultados válidos (status: %s), tentando fallback com BrasilAPI...", cep, geocodeResp.Status)
		goto TentarFallback
	}

	// 1.5: Sucesso na cadeia primária - combinar dados
	localidadeEndereco.Latitude = geocodeResp.Result[0].Geometry.Location.Lat
	localidadeEndereco.Longitude = geocodeResp.Result[0].Geometry.Location.Lng
	log.Printf("[Geo] CEP %s: Sucesso com cadeia primária! Coordenadas: lat=%.6f, lng=%.6f",
		cep, localidadeEndereco.Latitude, localidadeEndereco.Longitude)
	return localidadeEndereco, nil

TentarFallback:
	// ETAPA 2: Tentativa de Fallback - BrasilAPI Direta
	log.Printf("[Geo] CEP %s: Tentando fallback direto com BrasilAPI...", cep)

	localidadeBrasilAPI, err := s.brasilAPIClient.GetCEPInfo(cep)
	if err != nil {
		// ETAPA 3: Erro Final - todas as tentativas falharam
		log.Printf("[Geo] CEP %s: Falha no fallback BrasilAPI: %v", cep, err)
		return nil, fmt.Errorf("não foi possível obter dados de localização para o CEP %s: cadeia primária e fallback falharam", cep)
	}

	// Sucesso no fallback
	log.Printf("[Geo] CEP %s: Sucesso com fallback BrasilAPI! Coordenadas: lat=%.6f, lng=%.6f",
		cep, localidadeBrasilAPI.Latitude, localidadeBrasilAPI.Longitude)
	return localidadeBrasilAPI, nil
}
