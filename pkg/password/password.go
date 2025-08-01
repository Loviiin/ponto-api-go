package password

import (
	"golang.org/x/crypto/bcrypt"
)

func CriptografaSenha(senha string) (string, error) {
	senhaCripto, err := bcrypt.GenerateFromPassword([]byte(senha), 14)
	return string(senhaCripto), err
}

func VerificaHashSenha(senha string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha))
	return err == nil
}
