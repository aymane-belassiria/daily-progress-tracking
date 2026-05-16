package app

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

type tokenPayload struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
}

func issueToken(email, secret string) (string, error) {
	headerBytes, _ := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	payloadBytes, _ := json.Marshal(tokenPayload{
		Sub: email,
		Exp: time.Now().Add(7 * 24 * time.Hour).Unix(),
	})

	header := base64.RawURLEncoding.EncodeToString(headerBytes)
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := header + "." + payload
	signature := signToken(signingInput, secret)

	return signingInput + "." + signature, nil
}

func verifyToken(token, secret string) (tokenPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return tokenPayload{}, errors.New("invalid token format")
	}

	signingInput := parts[0] + "." + parts[1]
	expected := signToken(signingInput, secret)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[2])) != 1 {
		return tokenPayload{}, errors.New("invalid token signature")
	}

	rawPayload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return tokenPayload{}, errors.New("invalid token payload")
	}

	var payload tokenPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return tokenPayload{}, errors.New("invalid token payload")
	}
	if payload.Sub == "" || time.Now().Unix() > payload.Exp {
		return tokenPayload{}, errors.New("expired token")
	}

	return payload, nil
}

func signToken(input, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func secureStringEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
