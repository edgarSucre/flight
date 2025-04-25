package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func NewClaims(username string, duration time.Duration) (jwt.RegisteredClaims, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return jwt.RegisteredClaims{}, fmt.Errorf("unable to generate uuid: %w", err)
	}

	return jwt.RegisteredClaims{
		Issuer:    "test",
		Subject:   username,
		Audience:  []string{"test"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        tokenID.String(),
	}, nil
}
