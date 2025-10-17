package geolocation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/cep"
)

// Structs para a resposta do OpenCage
type OpenCageResponse struct {
	Results []struct {
		Geometry struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"geometry"`
	} `json:"results"`
}

type Service interface {
	GetLocationFromCEP(cep string) (*model.Localidade, error)
}

type service struct {
	httpClient     *http.Client
	openCageAPIKey string
	cepService     cep.Service
}

func NewService(apiKey string, cepService cep.Service) Service {
	return &service{
		httpClient:     &http.Client{},
		openCageAPIKey: apiKey,
		cepService:     cepService,
	}
}

func (s *service) GetLocationFromCEP(cep string) (*model.Localidade, error) {
	// Busca informações do CEP usando o serviço (BrasilAPI + fallback ViaCEP)
	localidade, err := s.cepService.GetAddressByCEP(cep)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar informações do CEP: %w", err)
	}

       // Se a BrasilAPI já retornou com coordenadas, não precisa chamar OpenCage
       if localidade.Latitude != 0 && localidade.Longitude != 0 {
	       fmt.Printf("[Geo] CEP %s: latitude/longitude da BrasilAPI: lat=%.6f, long=%.6f\n", cep, localidade.Latitude, localidade.Longitude)
	       return localidade, nil
       }

       // Se não tem coordenadas (veio do ViaCEP), busca via OpenCage
       if err := s.enrichWithOpenCage(localidade); err != nil {
	       fmt.Printf("[Geo] CEP %s: ViaCEP não retornou coordenadas e OpenCage falhou: %v\n", cep, err)
	       return nil, fmt.Errorf("falha ao obter coordenadas: %w", err)
       }

       if localidade.Latitude != 0 && localidade.Longitude != 0 {
	       fmt.Printf("[Geo] CEP %s: latitude/longitude do OpenCage (ViaCEP): lat=%.6f, long=%.6f\n", cep, localidade.Latitude, localidade.Longitude)
       } else {
	       fmt.Printf("[Geo] CEP %s: Nenhuma coordenada encontrada após ViaCEP e OpenCage.\n", cep)
       }
       return localidade, nil
}

// enrichWithOpenCage adiciona coordenadas à localidade usando OpenCage API
func (s *service) enrichWithOpenCage(localidade *model.Localidade) error {
	// Montar o endereço para consulta
	address := fmt.Sprintf("%s, %s, %s, %s, Brasil",
		localidade.Logradouro, localidade.Bairro, localidade.Cidade, localidade.Estado)

	openCageURL := fmt.Sprintf("https://api.opencagedata.com/geocode/v1/json?q=%s&key=%s",
		url.QueryEscape(address), s.openCageAPIKey)

	resp, err := s.httpClient.Get(openCageURL)
	if err != nil {
		return fmt.Errorf("falha ao realizar requisição para OpenCage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OpenCage retornou status %d", resp.StatusCode)
	}

	var openCageData OpenCageResponse
	if err := json.NewDecoder(resp.Body).Decode(&openCageData); err != nil {
		return fmt.Errorf("falha ao decodificar resposta da OpenCage: %w", err)
	}

	if len(openCageData.Results) == 0 {
		return fmt.Errorf("não foi possível obter as coordenadas para o endereço: %s", address)
	}

	// Atualiza as coordenadas na localidade existente
	localidade.Latitude = openCageData.Results[0].Geometry.Lat
	localidade.Longitude = openCageData.Results[0].Geometry.Lng

	return nil
}
