package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Claims struct {
	Subject   string `json:"sub"`
	ID        string `json:"jti"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func secret() ([]byte, error) {
	s := os.Getenv("JWT_SECRET")
	if len(s) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters")
	}
	return []byte(s), nil
}
func encode(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }
func signature(data string, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return encode(h.Sum(nil))
}
func Issue(userID uint) (string, Claims, error) {
	key, err := secret()
	if err != nil {
		return "", Claims{}, err
	}
	nonce := make([]byte, 16)
	if _, err = rand.Read(nonce); err != nil {
		return "", Claims{}, err
	}
	now := time.Now().Unix()
	c := Claims{Subject: strconv.FormatUint(uint64(userID), 10), ID: encode(nonce), IssuedAt: now, ExpiresAt: now + 3600}
	body, _ := json.Marshal(c)
	data := encode([]byte(`{"alg":"HS256","typ":"JWT"}`)) + "." + encode(body)
	return data + "." + signature(data, key), c, nil
}
func Verify(token string) (uint, Claims, error) {
	key, err := secret()
	if err != nil {
		return 0, Claims{}, err
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, Claims{}, errors.New("malformed token")
	}
	data := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(signature(data, key)), []byte(parts[2])) {
		return 0, Claims{}, errors.New("invalid signature")
	}
	head, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, Claims{}, err
	}
	var hdr struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if json.Unmarshal(head, &hdr) != nil || hdr.Alg != "HS256" || hdr.Typ != "JWT" {
		return 0, Claims{}, errors.New("invalid algorithm")
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, Claims{}, err
	}
	var c Claims
	if err = json.Unmarshal(body, &c); err != nil {
		return 0, Claims{}, err
	}
	now := time.Now().Unix()
	if c.ExpiresAt <= now || c.IssuedAt > now+60 || c.ID == "" {
		return 0, Claims{}, errors.New("expired or invalid token")
	}
	id, err := strconv.ParseUint(c.Subject, 10, 64)
	if err != nil || id == 0 {
		return 0, Claims{}, fmt.Errorf("invalid subject")
	}
	return uint(id), c, nil
}
