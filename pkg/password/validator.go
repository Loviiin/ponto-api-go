package password

import (
	"errors"
	"strings"
	"unicode"
)

// ValidarForcaSenha valida se a senha atende aos requisitos mínimos de segurança
// Requisitos:
// - Mínimo 8 caracteres
// - Pelo menos 1 letra maiúscula
// - Pelo menos 1 letra minúscula
// - Pelo menos 1 número
// - Pelo menos 1 caractere especial
func ValidarForcaSenha(senha string) error {
	if len(senha) < 8 {
		return errors.New("senha deve ter no mínimo 8 caracteres")
	}

	var (
		temMaiuscula bool
		temMinuscula bool
		temNumero    bool
		temEspecial  bool
	)

	for _, char := range senha {
		switch {
		case unicode.IsUpper(char):
			temMaiuscula = true
		case unicode.IsLower(char):
			temMinuscula = true
		case unicode.IsNumber(char):
			temNumero = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			temEspecial = true
		}
	}

	if !temMaiuscula {
		return errors.New("senha deve conter pelo menos 1 letra maiúscula")
	}

	if !temMinuscula {
		return errors.New("senha deve conter pelo menos 1 letra minúscula")
	}

	if !temNumero {
		return errors.New("senha deve conter pelo menos 1 número")
	}

	if !temEspecial {
		return errors.New("senha deve conter pelo menos 1 caractere especial")
	}

	return nil
}

// NormalizarEmail normaliza o email removendo espaços e convertendo para lowercase
// Isso previne problemas de duplicação e login case-sensitive
func NormalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
