package token

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/edgarSucre/flight/util"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidToken     = errors.New("token is invalid")
	ErrInvalidSecretKey = errors.New("secret key is too short")
)

const (

	// TODO remove all references to secret key, including .env
	minSecretKeySize = 32
)

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (*JWTMaker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, ErrInvalidSecretKey
	}

	return &JWTMaker{secretKey}, nil
}

func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	claims, err := NewClaims(username, duration)
	if err != nil {
		return "", fmt.Errorf("unable to create token claims: %w", err)
	}

	// jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// return jwtToken.SignedString([]byte(maker.secretKey))

	key, err := getPrivateKey()
	if err != nil {
		return "", fmt.Errorf("unable to load certificate: %w", err)
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)

	return jwtToken.SignedString(key)
}

func (maker *JWTMaker) VerifyToken(token string) (string, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		// if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		// 	return "", ErrInvalidToken
		// }

		// return []byte(maker.secretKey), nil

		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return "", ErrInvalidToken
		}

		return getPublicKey()
	}

	jwtToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrExpiredToken
		}

		return "", ErrInvalidToken
	}

	claims, ok := jwtToken.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	return claims.Subject, nil
}

func getPrivateKey() (any, error) {
	payload, err := os.ReadFile(util.FilePath("server.key"))
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(payload)

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	return key, nil
}

func getPublicKey() (any, error) {
	payload, err := os.ReadFile(util.FilePath("server.crt"))
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(payload)

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	key := cert.PublicKey.(*rsa.PublicKey)

	return key, nil
}
