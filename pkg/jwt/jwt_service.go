package jwt

import (
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
