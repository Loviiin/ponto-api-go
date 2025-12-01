package validator

import (
	"testing"
)

func TestValidarCPF(t *testing.T) {
	tests := []struct {
		name    string
		cpf     string
		wantErr bool
	}{
		{
			name:    "CPF Válido",
			cpf:     "52998224725", // CPF gerado para teste
			wantErr: false,
		},
		{
			name:    "CPF Válido com formatação",
			cpf:     "529.982.247-25",
			wantErr: false,
		},
		{
			name:    "CPF Inválido - Tamanho incorreto",
			cpf:     "1234567890",
			wantErr: true,
		},
		{
			name:    "CPF Inválido - Todos dígitos iguais",
			cpf:     "11111111111",
			wantErr: true,
		},
		{
			name:    "CPF Inválido - Dígito verificador incorreto",
			cpf:     "52998224726",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidarCPF(tt.cpf); (err != nil) != tt.wantErr {
				t.Errorf("ValidarCPF() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidarCNPJ(t *testing.T) {
	tests := []struct {
		name    string
		cnpj    string
		wantErr bool
	}{
		{
			name:    "CNPJ Válido",
			cnpj:    "00000000000191", // Banco do Brasil
			wantErr: false,
		},
		{
			name:    "CNPJ Válido com formatação",
			cnpj:    "00.000.000/0001-91",
			wantErr: false,
		},
		{
			name:    "CNPJ Inválido - Tamanho incorreto",
			cnpj:    "1234567890123",
			wantErr: true,
		},
		{
			name:    "CNPJ Inválido - Todos dígitos iguais",
			cnpj:    "11111111111111",
			wantErr: true,
		},
		{
			name:    "CNPJ Inválido - Dígito verificador incorreto",
			cnpj:    "37207431000161",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidarCNPJ(tt.cnpj); (err != nil) != tt.wantErr {
				t.Errorf("ValidarCNPJ() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidarTelefoneBrasileiro(t *testing.T) {
	tests := []struct {
		name     string
		telefone string
		wantErr  bool
	}{
		{
			name:     "Celular Válido",
			telefone: "11999999999",
			wantErr:  false,
		},
		{
			name:     "Celular Válido com formatação",
			telefone: "(11) 99999-9999",
			wantErr:  false,
		},
		{
			name:     "DDD Inválido",
			telefone: "01999999999",
			wantErr:  true,
		},
		{
			name:     "Celular não começa com 9",
			telefone: "11899999999",
			wantErr:  true,
		},
		{
			name:     "Tamanho incorreto",
			telefone: "1199999999",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidarTelefoneBrasileiro(tt.telefone); (err != nil) != tt.wantErr {
				t.Errorf("ValidarTelefoneBrasileiro() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
