package geolocation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Loviiin/ponto-api-go/internal/model"
)

// Struct para a resposta do ViaCEP
type ViaCEPResponse struct {
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
}

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
}

func NewService(apiKey string) Service {
	return &service{
		httpClient:     &http.Client{},
		openCageAPIKey: apiKey,
	}
}

func (s *service) GetLocationFromCEP(cep string) (*model.Localidade, error) {
	viaCEPURL := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	resp, err := s.httpClient.Get(viaCEPURL)
	if err != nil {
		return nil, fmt.Errorf("falha ao realizar requisição para ViaCEP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ViaCEP retornou status %d", resp.StatusCode)
	}

	var viaCEPData ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&viaCEPData); err != nil {
		return nil, fmt.Errorf("falha ao decodificar resposta da ViaCEP: %w", err)
	}

	// 2. Montar o endereço e chamar OpenCage
	address := fmt.Sprintf("%s, %s, %s, %s, Brasil",
		viaCEPData.Logradouro, viaCEPData.Bairro, viaCEPData.Localidade, viaCEPData.UF)

	openCageURL := fmt.Sprintf("https://api.opencagedata.com/geocode/v1/json?q=%s&key=%s",
		url.QueryEscape(address), s.openCageAPIKey)

	resp, err = s.httpClient.Get(openCageURL)
	if err != nil {
		return nil, fmt.Errorf("falha ao realizar requisição para OpenCage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenCage retornou status %d", resp.StatusCode)
	}

	var openCageData OpenCageResponse
	if err := json.NewDecoder(resp.Body).Decode(&openCageData); err != nil {
		return nil, fmt.Errorf("falha ao decodificar resposta da OpenCage: %w", err)
	}

	if len(openCageData.Results) == 0 {
		return nil, fmt.Errorf("não foi possível obter as coordenadas para o endereço: %s", address)
	}

	// 3. Montar e retornar o objeto Localidade
	localidade := &model.Localidade{
		CEP:        cep,
		Logradouro: viaCEPData.Logradouro,
		Bairro:     viaCEPData.Bairro,
		Cidade:     viaCEPData.Localidade,
		Estado:     viaCEPData.UF,
		Latitude:   openCageData.Results[0].Geometry.Lat,
		Longitude:  openCageData.Results[0].Geometry.Lng,
	}

	return localidade, nil
}
