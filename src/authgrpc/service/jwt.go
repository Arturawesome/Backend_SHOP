package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const passwordAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type JWTProvider struct {
	secretKey []byte
}

type UserClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewJWTProvider(secret string) *JWTProvider {
	return &JWTProvider{secretKey: []byte(secret)}
}

func (provider *JWTProvider) GenerateAccessToken(
	userID uuid.UUID) (string, error) {
	claims := UserClaims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString(provider.secretKey)
}

func (provider *JWTProvider) ValidateTokenAndGetUserID(tokenString string,
) (string, error) {

	if tokenString == "" {
		return "", ErrInvalidToken
	}
	claims, err := provider.parseToken(tokenString)
	if err != nil {
		return "", ErrInvalidToken
	}

	return claims.UserID, nil
}

func (provider *JWTProvider) parseToken(tokenString string,
) (*UserClaims, error) {

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}

		return provider.secretKey, nil
	}

	claims := &UserClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		keyFunc,
	)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	if claims.UserID == "" {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func generatePassword(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("password length must be greater than zero")
	}

	password := make([]byte, length)

	for i := range password {
		index, err := rand.Int(
			rand.Reader,
			big.NewInt(int64(len(passwordAlphabet))),
		)
		if err != nil {
			return "", err
		}

		password[i] = passwordAlphabet[index.Int64()]
	}

	return string(password), nil
}
