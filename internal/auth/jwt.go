package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrTokenInvalido = errors.New("token invalido")

type Claims struct {
	UserID string `json:"sub"`
	Correo string `json:"correo"`
	Rol    string `json:"rol"`
	Issuer string `json:"iss"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

func GenerateToken(secret string, issuer string, duration time.Duration, userID string, correo string, rol string) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(duration)

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	claims := Claims{
		UserID: userID,
		Correo: correo,
		Rol:    rol,
		Issuer: issuer,
		Exp:    expiresAt.Unix(),
		Iat:    now.Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", time.Time{}, err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, err
	}

	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	signature := sign(unsigned, secret)

	return unsigned + "." + signature, expiresAt, nil
}

func ParseToken(token string, secret string, issuer string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrTokenInvalido
	}

	unsigned := parts[0] + "." + parts[1]
	expectedSignature := sign(unsigned, secret)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return nil, ErrTokenInvalido
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrTokenInvalido
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrTokenInvalido
	}

	if claims.Issuer != issuer || claims.Exp <= time.Now().UTC().Unix() || claims.UserID == "" || claims.Rol == "" {
		return nil, ErrTokenInvalido
	}

	return &claims, nil
}

func sign(value string, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func BearerToken(authorizationHeader string) (string, error) {
	parts := strings.Fields(authorizationHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("%w: authorization bearer requerido", ErrTokenInvalido)
	}
	return parts[1], nil
}
