package viacep

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Loviiin/ponto-api-go/internal/model"
)

type ViaCEPResponse struct {
	CEP         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	UF          string `json:"uf"`
	IBGE        string `json:"ibge"`
	GIA         string `json:"gia"`
	DDD         string `json:"ddd"`
	SIAFI       string `json:"siafi"`
	Erro        bool   `json:"erro,omitempty"`
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
	viaCEPURL := fmt.Sprintf("%s/%s/json/", c.urlApi, cep)

	resposta, err := c.httpClient.Get(viaCEPURL)
	if err != nil {
		return nil, fmt.Errorf("falha ao realizar requisição para ViaCEP: %w", err)
	}
	defer resposta.Body.Close()

	if resposta.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ViaCEP retornou status: %d", resposta.StatusCode)
	}

	var viaCEPData ViaCEPResponse
	if err := json.NewDecoder(resposta.Body).Decode(&viaCEPData); err != nil {
		return nil, fmt.Errorf("falha ao decodificar resposta do ViaCEP: %w", err)
	}

	// ViaCEP retorna um campo "erro": true quando o CEP não existe
	if viaCEPData.Erro {
		return nil, fmt.Errorf("CEP não encontrado no ViaCEP")
	}

	// Converter a resposta do ViaCEP para o modelo Localidade
	localidade := &model.Localidade{
		CEP:        viaCEPData.CEP,
		Logradouro: viaCEPData.Logradouro,
		Bairro:     viaCEPData.Bairro,
		Cidade:     viaCEPData.Localidade,
		Estado:     viaCEPData.UF,
		// Nota: ViaCEP não retorna coordenadas
	}

	return localidade, nil
}
