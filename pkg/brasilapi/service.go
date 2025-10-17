package brasilapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Loviiin/ponto-api-go/internal/model"
)

type BrasilAPICEPResponse struct {
	CEP          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service,omitempty"`
	Location     *struct {
		Type        string `json:"type"`
		Coordinates *struct {
			Longitude string `json:"longitude"`
			Latitude  string `json:"latitude"`
		} `json:"coordinates"`
	} `json:"location"`
}

type Client interface {
	GetCEPInfo(cep string) (*model.Localidade, error)
}

type client struct {
	httpClient *http.Client
	urlApi     string
}

func NewClient(urlApi string) Client {
	return &client{
		httpClient: &http.Client{},
		urlApi:     urlApi,
	}
}

func (c *client) GetCEPInfo(cep string) (*model.Localidade, error) {
	brasilCEPURL := fmt.Sprintf("%s/cep/v2/%s", c.urlApi, cep)

	resposta, err := c.httpClient.Get(brasilCEPURL)
	if err != nil {
		return nil, fmt.Errorf("falha ao realizar requisição para BrasilAPI: %w", err)
	}
	defer resposta.Body.Close()

	if resposta.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("BrasilAPI retornou status: %d", resposta.StatusCode)
	}

	var brasilAPIData BrasilAPICEPResponse
	if err := json.NewDecoder(resposta.Body).Decode(&brasilAPIData); err != nil {
		return nil, fmt.Errorf("falha ao decodificar resposta da BrasilAPI: %w", err)
	}

	// Converter a resposta da BrasilAPI para o modelo Localidade
	localidade := &model.Localidade{
		CEP:        brasilAPIData.CEP,
		Logradouro: brasilAPIData.Street,
		Bairro:     brasilAPIData.Neighborhood,
		Cidade:     brasilAPIData.City,
		Estado:     brasilAPIData.State,
	}

	// Converter coordenadas se disponíveis
	if brasilAPIData.Location != nil && brasilAPIData.Location.Coordinates != nil {
		// BrasilAPI retorna strings; tente converter para float64
		if latStr := brasilAPIData.Location.Coordinates.Latitude; latStr != "" {
			if lngStr := brasilAPIData.Location.Coordinates.Longitude; lngStr != "" {
				if lat, errLat := strconv.ParseFloat(latStr, 64); errLat == nil {
					if lng, errLng := strconv.ParseFloat(lngStr, 64); errLng == nil {
						localidade.Latitude = lat
						localidade.Longitude = lng
					}
				}
			}
		}
	}

	return localidade, nil
}
