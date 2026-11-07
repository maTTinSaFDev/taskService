package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"crypto/sha256"
	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenManager interface {
	Issue(userID int64) (string, time.Time, error)
	Verify(token string) (int64, error)
}

type JWTManager struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	issuer     string
	audience   string
	ttl        time.Duration
	now        func() time.Time
}

func NewJWTManager(privateKey ed25519.PrivateKey, issuer string, audience string, ttl time.Duration) (*JWTManager, error) {
	if len(privateKey) != ed25519.PrivateKeySize || issuer == "" || audience == "" || ttl <= 0 {
		return nil, errors.New("valid Ed25519 key, issuer, audience and access TTL are required")
	}
	publicKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("derive Ed25519 public key")
	}
	return &JWTManager{privateKey: privateKey, publicKey: publicKey, issuer: issuer, audience: audience, ttl: ttl, now: time.Now}, nil
}

func (manager *JWTManager) Issue(userID int64) (string, time.Time, error) {
	if userID <= 0 {
		return "", time.Time{}, errors.New("user id must be greater than zero")
	}
	now := manager.now().UTC()
	expiresAt := now.Add(manager.ttl)
	tokenID, err := randomID()
	if err != nil {
		return "", time.Time{}, err
	}
	claims := jwt.RegisteredClaims{
		Issuer:    manager.issuer,
		Subject:   strconv.FormatInt(userID, 10),
		Audience:  jwt.ClaimStrings{manager.audience},
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ID:        tokenID,
	}
	value, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(manager.privateKey)
	return value, expiresAt, err
}

func (manager *JWTManager) Verify(value string) (int64, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(value, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodEdDSA {
			return nil, errors.New("unexpected JWT signing method")
		}
		return manager.publicKey, nil
	}, jwt.WithIssuer(manager.issuer), jwt.WithAudience(manager.audience), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithTimeFunc(manager.now))
	if err != nil || !token.Valid {
		return 0, errors.New("invalid access token")
	}
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID <= 0 {
		return 0, errors.New("invalid access token subject")
	}
	return userID, nil
}

func LoadEd25519PrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read JWT private key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("JWT private key is not PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse JWT private key: %w", err)
	}
	privateKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("JWT private key is not Ed25519")
	}
	return privateKey, nil
}

func newRefreshToken() (raw string, hash string, err error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(buffer)
	sum := sha256.Sum256([]byte(raw))
	return raw, hex.EncodeToString(sum[:]), nil
}

func hashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func randomID() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}
