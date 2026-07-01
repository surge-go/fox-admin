package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type tokenHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

// SignAccess 签发 access token。
func (m *Manager) SignAccess(claims Claims) (string, error) {
	header, err := encodeTokenPart(tokenHeader{Algorithm: "HS256", Type: "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := encodeTokenPart(claims)
	if err != nil {
		return "", err
	}
	unsigned := header + "." + payload
	signature := m.sign(unsigned)
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (m *Manager) parseAccess(token string) (*Claims, error) {
	token, err := ParseBearer(token)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrTokenMalformed
	}

	headerPayload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrTokenMalformed
	}
	var header tokenHeader
	if err := json.Unmarshal(headerPayload, &header); err != nil {
		return nil, ErrTokenMalformed
	}
	if header.Algorithm != "HS256" || header.Type != "JWT" {
		return nil, ErrTokenMalformed
	}

	unsigned := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrTokenMalformed
	}
	expected := m.sign(unsigned)
	if subtle.ConstantTimeCompare(signature, expected) != 1 {
		return nil, ErrInvalidSignature
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrTokenMalformed
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrTokenMalformed
	}
	if err := claims.validate(); err != nil {
		return nil, err
	}
	if m.cfg.Issuer != "" && claims.Issuer != m.cfg.Issuer {
		return nil, ErrTokenMalformed
	}
	if m.cfg.Audience != "" && claims.Audience != m.cfg.Audience {
		return nil, ErrTokenMalformed
	}
	if !m.cfg.now().Before(claims.ExpiresAt) {
		return nil, ErrTokenExpired
	}
	return &claims, nil
}

func (claims Claims) validate() error {
	if claims.SubjectID <= 0 || strings.TrimSpace(string(claims.SubjectType)) == "" {
		return ErrTokenMalformed
	}
	if strings.TrimSpace(string(claims.Provider)) == "" {
		return ErrTokenMalformed
	}
	if strings.TrimSpace(string(claims.Platform)) == "" {
		return ErrTokenMalformed
	}
	if strings.TrimSpace(claims.SessionID) == "" || strings.TrimSpace(claims.TokenID) == "" {
		return ErrTokenMalformed
	}
	if claims.IssuedAt.IsZero() || claims.ExpiresAt.IsZero() {
		return ErrTokenMalformed
	}
	if !claims.ExpiresAt.After(claims.IssuedAt.Add(-time.Second)) {
		return ErrTokenMalformed
	}
	return nil
}

func (m *Manager) sign(value string) []byte {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}

func encodeTokenPart(value any) (string, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func mapTokenError(err error) error {
	if errors.Is(err, ErrTokenRequired) ||
		errors.Is(err, ErrTokenMalformed) ||
		errors.Is(err, ErrTokenExpired) ||
		errors.Is(err, ErrInvalidSignature) {
		return err
	}
	return ErrTokenMalformed
}
