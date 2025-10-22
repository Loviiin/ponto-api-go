package distancematrix

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Response models based on the provided JSON example
// {
//   "result": [ { ... } ],
//   "status": "OK"
// }

type AddressComponent struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Viewport struct {
	Northeast LatLng `json:"northeast"`
	Southwest LatLng `json:"southwest"`
}

type Geometry struct {
	Location     LatLng   `json:"location"`
	LocationType string   `json:"location_type"`
	Viewport     Viewport `json:"viewport"`
}

type GeocodeResult struct {
	AddressComponents []AddressComponent `json:"address_components"`
	FormattedAddress  string             `json:"formatted_address"`
	Geometry          Geometry           `json:"geometry"`
	PlaceID           string             `json:"place_id"`
	PlusCode          map[string]any     `json:"plus_code"`
	Types             []string           `json:"types"`
}

type GeocodeResponse struct {
	Result []GeocodeResult `json:"result"`
	Status string          `json:"status"`
}


type Client interface {
	GeocodeAddress(address string) (*GeocodeResponse, error)
}

type client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

const defaultBaseURL = "https://api.distancematrix.ai/maps/api/geocode/json"


func NewClient(apiKey string) Client {
	return &client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
	}
}

func (c *client) GeocodeAddress(address string) (*GeocodeResponse, error) {
	if address == "" {
		return nil, fmt.Errorf("endereço vazio")
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("URL base inválida: %w", err)
	}

	q := u.Query()
	q.Set("address", address)
	q.Set("key", c.apiKey)
	q.Set("language", "pt")
	q.Set("region", "br")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar requisição: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao realizar requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API retornou status %d", resp.StatusCode)
	}

	var out GeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("falha ao decodificar JSON: %w", err)
	}

	return &out, nil
}
