package validator

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// SanitizarCPF remove caracteres não numéricos do CPF
func SanitizarCPF(cpf string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, cpf)
}

// SanitizarCNPJ remove caracteres não numéricos do CNPJ
func SanitizarCNPJ(cnpj string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, cnpj)
}

// SanitizarTelefone remove caracteres não numéricos do telefone
func SanitizarTelefone(telefone string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, telefone)
}

// ValidarCPF valida se o CPF é válido segundo as regras brasileiras
func ValidarCPF(cpf string) error {
	// Sanitizar primeiro
	cpf = SanitizarCPF(cpf)

	// CPF deve ter 11 dígitos
	if len(cpf) != 11 {
		return errors.New("CPF deve conter 11 dígitos")
	}

	// Verificar se todos os dígitos são iguais (CPF inválido)
	allSame := true
	for i := 1; i < len(cpf); i++ {
		if cpf[i] != cpf[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return errors.New("CPF inválido: todos os dígitos são iguais")
	}

	// Validar primeiro dígito verificador
	sum := 0
	for i := 0; i < 9; i++ {
		digit, _ := strconv.Atoi(string(cpf[i]))
		sum += digit * (10 - i)
	}
	remainder := sum % 11
	firstDigit := 0
	if remainder >= 2 {
		firstDigit = 11 - remainder
	}

	firstCheckDigit, _ := strconv.Atoi(string(cpf[9]))
	if firstDigit != firstCheckDigit {
		return errors.New("CPF inválido: dígito verificador incorreto")
	}

	// Validar segundo dígito verificador
	sum = 0
	for i := 0; i < 10; i++ {
		digit, _ := strconv.Atoi(string(cpf[i]))
		sum += digit * (11 - i)
	}
	remainder = sum % 11
	secondDigit := 0
	if remainder >= 2 {
		secondDigit = 11 - remainder
	}

	secondCheckDigit, _ := strconv.Atoi(string(cpf[10]))
	if secondDigit != secondCheckDigit {
		return errors.New("CPF inválido: dígito verificador incorreto")
	}

	return nil
}

// ValidarCNPJ valida se o CNPJ é válido segundo as regras brasileiras
func ValidarCNPJ(cnpj string) error {
	// Sanitizar primeiro
	cnpj = SanitizarCNPJ(cnpj)

	// CNPJ deve ter 14 dígitos
	if len(cnpj) != 14 {
		return errors.New("CNPJ deve conter 14 dígitos")
	}

	// Verificar se todos os dígitos são iguais (CNPJ inválido)
	allSame := true
	for i := 1; i < len(cnpj); i++ {
		if cnpj[i] != cnpj[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return errors.New("CNPJ inválido: todos os dígitos são iguais")
	}

	// Validar primeiro dígito verificador
	weights1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 12; i++ {
		digit, _ := strconv.Atoi(string(cnpj[i]))
		sum += digit * weights1[i]
	}
	remainder := sum % 11
	firstDigit := 0
	if remainder >= 2 {
		firstDigit = 11 - remainder
	}

	firstCheckDigit, _ := strconv.Atoi(string(cnpj[12]))
	if firstDigit != firstCheckDigit {
		return errors.New("CNPJ inválido: dígito verificador incorreto")
	}

	// Validar segundo dígito verificador
	weights2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum = 0
	for i := 0; i < 13; i++ {
		digit, _ := strconv.Atoi(string(cnpj[i]))
		sum += digit * weights2[i]
	}
	remainder = sum % 11
	secondDigit := 0
	if remainder >= 2 {
		secondDigit = 11 - remainder
	}

	secondCheckDigit, _ := strconv.Atoi(string(cnpj[13]))
	if secondDigit != secondCheckDigit {
		return errors.New("CNPJ inválido: dígito verificador incorreto")
	}

	return nil
}

// ValidarTelefoneBrasileiro valida formato de telefone brasileiro
// Aceita: (XX) XXXXX-XXXX (celular com 9 dígitos)
func ValidarTelefoneBrasileiro(telefone string) error {
	// Sanitizar primeiro
	sanitized := SanitizarTelefone(telefone)

	// Telefone brasileiro deve ter 11 dígitos (DDD + 9 dígitos do celular)
	if len(sanitized) != 11 {
		return errors.New("telefone deve ter 11 dígitos no formato (XX) XXXXX-XXXX")
	}

	// Validar DDD (11 a 99)
	ddd, err := strconv.Atoi(sanitized[0:2])
	if err != nil || ddd < 11 || ddd > 99 {
		return errors.New("DDD inválido (deve ser entre 11 e 99)")
	}

	// Celular deve começar com 9
	if sanitized[2] != '9' {
		return errors.New("número de celular deve começar com 9")
	}

	// Validar formato com regex (opcional, para quando vier formatado)
	if telefone != sanitized {
		matched, _ := regexp.MatchString(`^\(\d{2}\)\s?\d{5}-?\d{4}$`, telefone)
		if !matched {
			return errors.New("formato inválido. Use: (XX) XXXXX-XXXX")
		}
	}

	return nil
}
