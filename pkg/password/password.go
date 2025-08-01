package password

import (
	"golang.org/x/crypto/bcrypt"
)

func CriptografaSenha(senha string) (string, error) {
	senhaCripto, err := bcrypt.GenerateFromPassword([]byte(senha), 14)
	return string(senhaCripto), err
}

func VerificaHashSenha(senha string, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha))
	if err != nil {
		return false, err
	}
	return true, nil
}
