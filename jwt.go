package flight

import (
	"time"

	"github.com/google/uuid"
)

type JwtPayload struct {
	ExpiredAt time.Time
	ID        uuid.UUID
	IssuedAt  time.Time
	Username  string
}

// func NewPayload(username string, duration time.Duration) (*JwtPayload, error) {
// 	tokenID, err := uuid.NewRandom()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &JwtPayload{
// 		ExpiredAt: time.Now().Add(duration),
// 		ID:        tokenID,
// 		IssuedAt:  time.Now(),
// 		Username:  username,
// 	}, nil
// }

// type JwtManager interface {
// 	CreateToken(username string, duration time.Duration) (string, error)
// 	VerifyToken(token string) (*JwtPayload, error)
// }
