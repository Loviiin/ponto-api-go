package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"

	"strconv"
	"time"
)

type JWTService struct {
	secretKey string
	issuer    string
}

func NewJWTService(secretKey string, issuer string) *JWTService {
	return &JWTService{secretKey: secretKey, issuer: issuer}
}

func (service *JWTService) GenerateToken(userID uint) (string, error) {

	jwtClaim := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    service.issuer,
		Subject:   strconv.Itoa(int(userID)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaim)
	tokenString, err := token.SignedString([]byte(service.secretKey))

	return tokenString, err
}

func (s *JWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})
}
