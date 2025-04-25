package token_test

import (
	"testing"
	"time"

	"github.com/edgarSucre/flight/token"
	"github.com/edgarSucre/flight/util"
	"github.com/stretchr/testify/assert"
)

func TestJWTMaker(t *testing.T) {
	maker, err := token.NewJWTMaker(util.RandomString(32))
	if err != nil {
		t.Fatal(err)
	}

	username := util.RandomString(6)

	validToken, err := maker.CreateToken(username, time.Minute)
	assert.NoError(t, err)
	assert.NotEmpty(t, validToken)

	expiredToken, err := maker.CreateToken(username, time.Nanosecond)
	assert.NoError(t, err)
	assert.NotEmpty(t, expiredToken)

	time.Sleep(time.Millisecond)

	tests := []struct {
		name         string
		tokenPayload string
		err          error
	}{
		{
			"validToken",
			validToken,
			nil,
		},
		{
			"expiredToken",
			expiredToken,
			token.ErrExpiredToken,
		},
		{
			"invalidToken",
			"not a token at all",
			token.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := maker.VerifyToken(tt.tokenPayload)

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)
				assert.Empty(t, actual)
			} else {
				assert.Equal(t, username, actual)
			}

		})
	}
}

func TestNewJWTMaker(t *testing.T) {
	maker, err := token.NewJWTMaker(util.RandomString(32))
	assert.NoError(t, err)
	assert.NotNil(t, maker)

	maker, err = token.NewJWTMaker("short key")
	assert.ErrorIs(t, err, token.ErrInvalidSecretKey)
}
