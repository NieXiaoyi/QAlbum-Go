package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret     []byte
	expireHour int
}

func NewJWTManager(secret string, expireHour int) *JWTManager {
	return &JWTManager{
		secret:     []byte(secret),
		expireHour: expireHour,
	}
}

func (j *JWTManager) GenerateToken(userID int) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(j.expireHour) * time.Hour)

	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}

func (j *JWTManager) ParseToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, fmt.Errorf("invalid token")
}
