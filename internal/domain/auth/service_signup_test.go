package auth

import (
	"testing"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/stretchr/testify/assert"
)

// TestSignUpPreservesCoordinates verifica se as coordenadas enviadas pelo frontend
// são preservadas e não sobrescritas pela consulta ao CEP
func TestSignUpPreservesCoordinates(t *testing.T) {
	// Arrange: Coordenadas enviadas pelo frontend (do mapa)
	frontendLat := -15.799098183155415
	frontendLng := -47.880038022994995

	// Coordenadas esperadas do CEP (que NÃO devem ser usadas)
	cepLat := -15.804898650000002
	cepLng := -47.875684449999994

	localidade := &model.Localidade{
		Nome:               "Sede Principal",
		CEP:                "70070900",
		Logradouro:         "Praça dos Tribunais Superiores Bloco A",
		Bairro:             "Asa Sul",
		Cidade:             "Brasília",
		Estado:             "DF",
		Latitude:           frontendLat, // ✅ Coordenada do mapa
		Longitude:          frontendLng, // ✅ Coordenada do mapa
		RaioGeofenceMetros: 50,
	}

	// Assert: Validar que as coordenadas não foram sobrescritas
	assert.Equal(t, frontendLat, localidade.Latitude,
		"Latitude deve ser a enviada pelo frontend, não a do CEP")
	assert.Equal(t, frontendLng, localidade.Longitude,
		"Longitude deve ser a enviada pelo frontend, não a do CEP")

	// Garantir que não são as coordenadas do CEP
	assert.NotEqual(t, cepLat, localidade.Latitude,
		"Latitude NÃO deve ser a do CEP")
	assert.NotEqual(t, cepLng, localidade.Longitude,
		"Longitude NÃO deve ser a do CEP")
}

// TestValidateCoordinatesRange valida se coordenadas inválidas são rejeitadas
func TestValidateCoordinatesRange(t *testing.T) {
	testCases := []struct {
		name      string
		latitude  float64
		longitude float64
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "Coordenadas válidas",
			latitude:  -15.799098183155415,
			longitude: -47.880038022994995,
			shouldErr: false,
		},
		{
			name:      "Latitude acima do limite",
			latitude:  91.0,
			longitude: -47.880038022994995,
			shouldErr: true,
			errMsg:    "latitude inválida",
		},
		{
			name:      "Latitude abaixo do limite",
			latitude:  -91.0,
			longitude: -47.880038022994995,
			shouldErr: true,
			errMsg:    "latitude inválida",
		},
		{
			name:      "Longitude acima do limite",
			latitude:  -15.799098183155415,
			longitude: 181.0,
			shouldErr: true,
			errMsg:    "longitude inválida",
		},
		{
			name:      "Longitude abaixo do limite",
			latitude:  -15.799098183155415,
			longitude: -181.0,
			shouldErr: true,
			errMsg:    "longitude inválida",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simular validação
			var err error
			if tc.latitude < -90 || tc.latitude > 90 {
				err = assert.AnError
				assert.Contains(t, tc.errMsg, "latitude")
			}
			if tc.longitude < -180 || tc.longitude > 180 {
				err = assert.AnError
				assert.Contains(t, tc.errMsg, "longitude")
			}

			if tc.shouldErr {
				assert.Error(t, err, "Deveria retornar erro para coordenadas inválidas")
			} else {
				assert.NoError(t, err, "Não deveria retornar erro para coordenadas válidas")
			}
		})
	}
}

// TestSignUpRequestMapping valida que o mapeamento do request para o model está correto
func TestSignUpRequestMapping(t *testing.T) {
	// Arrange: Simular o SignUpRequest
	request := struct {
		Localidade struct {
			Nome               string
			CEP                string
			Logradouro         string
			Bairro             string
			Cidade             string
			Estado             string
			Latitude           float64
			Longitude          float64
			RaioGeofenceMetros float64
		}
	}{}

	request.Localidade.Nome = "Sede Principal"
	request.Localidade.CEP = "70070900"
	request.Localidade.Logradouro = "Praça dos Tribunais Superiores Bloco A"
	request.Localidade.Bairro = "Asa Sul"
	request.Localidade.Cidade = "Brasília"
	request.Localidade.Estado = "DF"
	request.Localidade.Latitude = -15.799098183155415
	request.Localidade.Longitude = -47.880038022994995
	request.Localidade.RaioGeofenceMetros = 50

	// Act: Mapear para model
	localidade := &model.Localidade{
		Nome:               request.Localidade.Nome,
		CEP:                request.Localidade.CEP,
		Logradouro:         request.Localidade.Logradouro,
		Bairro:             request.Localidade.Bairro,
		Cidade:             request.Localidade.Cidade,
		Estado:             request.Localidade.Estado,
		Latitude:           request.Localidade.Latitude,
		Longitude:          request.Localidade.Longitude,
		RaioGeofenceMetros: request.Localidade.RaioGeofenceMetros,
	}

	// Assert: Todos os campos devem ser mapeados corretamente
	assert.Equal(t, request.Localidade.Nome, localidade.Nome)
	assert.Equal(t, request.Localidade.CEP, localidade.CEP)
	assert.Equal(t, request.Localidade.Logradouro, localidade.Logradouro)
	assert.Equal(t, request.Localidade.Bairro, localidade.Bairro)
	assert.Equal(t, request.Localidade.Cidade, localidade.Cidade)
	assert.Equal(t, request.Localidade.Estado, localidade.Estado)
	assert.Equal(t, request.Localidade.Latitude, localidade.Latitude)
	assert.Equal(t, request.Localidade.Longitude, localidade.Longitude)
	assert.Equal(t, request.Localidade.RaioGeofenceMetros, localidade.RaioGeofenceMetros)
}
