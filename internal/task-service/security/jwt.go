package security

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTVerifier struct {
	publicKey ed25519.PublicKey
	issuer    string
	audience  string
	now       func() time.Time
}

func NewJWTVerifier(publicKey ed25519.PublicKey, issuer string, audience string) (*JWTVerifier, error) {
	if len(publicKey) != ed25519.PublicKeySize || issuer == "" || audience == "" {
		return nil, errors.New("valid Ed25519 public key, issuer and audience are required")
	}
	return &JWTVerifier{publicKey: publicKey, issuer: issuer, audience: audience, now: time.Now}, nil
}

func (verifier *JWTVerifier) Verify(value string) (int64, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(value, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodEdDSA {
			return nil, errors.New("unexpected JWT signing method")
		}
		return verifier.publicKey, nil
	}, jwt.WithIssuer(verifier.issuer), jwt.WithAudience(verifier.audience), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithTimeFunc(verifier.now))
	if err != nil || !token.Valid {
		return 0, errors.New("invalid access token")
	}
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID <= 0 {
		return 0, errors.New("invalid access token subject")
	}
	return userID, nil
}

func LoadEd25519PublicKey(path string) (ed25519.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read JWT public key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("JWT public key is not PEM")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse JWT public key: %w", err)
	}
	publicKey, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("JWT public key is not Ed25519")
	}
	return publicKey, nil
}
