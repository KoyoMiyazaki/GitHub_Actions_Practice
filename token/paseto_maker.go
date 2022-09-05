package token

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

type CustomPasetoClaims struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	paseto.JSONToken
}

func NewPasetoMaker(symmetricKey string) (Maker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}
	return maker, nil
}

func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	now := time.Now()
	exp := now.Add(duration)

	jsonToken := paseto.JSONToken{
		IssuedAt:   now,
		Expiration: exp,
	}
	jsonToken.Set("id", tokenID.String())
	jsonToken.Set("username", username)

	return maker.paseto.Encrypt(maker.symmetricKey, jsonToken, nil)
}

func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	var payload paseto.JSONToken

	err := maker.paseto.Decrypt(token, maker.symmetricKey, &payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().After(payload.Expiration) {
		return nil, ErrExpiredToken
	}

	uuid, err := uuid.Parse(payload.Get("id"))
	if err != nil {
		return nil, ErrInvalidToken
	}
	return &Payload{
		ID:        uuid,
		Username:  payload.Get("username"),
		IssuedAt:  payload.IssuedAt,
		ExpiresAt: payload.Expiration,
	}, nil
}
