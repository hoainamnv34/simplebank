package token

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInValidToken = errors.New("Invalid Token")
	ErrExpiredToken = errors.New("Token has expired")
)

// Payload contains the payload data of the token
type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username`
	IssueAt   time.Time `json:"issued_at`
	ExpiredAt time.Time `json:"expired_at`
}

// GetIssuer implements jwt.Claims.
func (*Payload) GetIssuer() (string, error) {
	return "Issuer", nil
}

// GetNotBefore implements jwt.Claims.
func (*Payload) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now()), nil
}

// GetSubject implements jwt.Claims.
func (*Payload) GetSubject() (string, error) {
	return "GetSubject", nil
}

// GetAudience implements jwt.Claims.
func (*Payload) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{"GetAudience"}, nil
}

// GetExpirationTime implements jwt.Claims.
func (p *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	expirationTime := jwt.NewNumericDate(p.ExpiredAt)
	return expirationTime, nil
}

// GetIssuedAt implements jwt.Claims.
func (p *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	issued_at := jwt.NewNumericDate(p.IssueAt)
	return issued_at, nil
}

// NewPayload creates a new token payload with a specific username and duration
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssueAt:   time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, nil
}

// Valid checks if the token payload is valid or not
func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
