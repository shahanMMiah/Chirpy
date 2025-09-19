package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const AUTHKEY = "Authorization"
const CHIRPY = "Chirpy"
const BEARERPREF = "Bearer"
const POLKAPREF = "ApiKey"

func HashPassword(password string) (string, error) {
	hashData, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashData), nil

}

func CheckPasswordHash(password, hash string) error {

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

}

func MakeJWT(userId uuid.UUID, tokenSecret string, expires time.Duration) (string, error) {

	claims := jwt.RegisteredClaims{
		Issuer:    CHIRPY,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expires).UTC()),
		Subject:   userId.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))

}

func ValidateJWT(signedToken, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(signedToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	subject, err := token.Claims.GetSubject()

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("subject faulure:  %s", err.Error())
	}
	id, err := uuid.Parse(subject)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("uuid parse faulure:  %s", err.Error())
	}
	return id, nil
}

func GetBearerToken(h http.Header) (string, error) {
	tokenstring, err := GetHeaderKey(h, AUTHKEY, BEARERPREF)
	return tokenstring, err
}

func GetAPIKey(h http.Header) (string, error) {
	tokenstring, err := GetHeaderKey(h, AUTHKEY, POLKAPREF)
	return tokenstring, err

}

func GetHeaderKey(h http.Header, hkey, keyPefix string) (string, error) {
	headerKey := h.Get(hkey)
	if headerKey == "" {
		return "", fmt.Errorf("Could not find %s Header", AUTHKEY)
	}

	tokenstring, found := strings.CutPrefix(headerKey, fmt.Sprintf("%s ", keyPefix))

	if !found {
		return "", fmt.Errorf("header prefix not found in header %s", headerKey)
	}

	return tokenstring, nil

}

func MakeRefreshToken() (string, error) {
	randBytes := make([]byte, 32)
	rand.Read(randBytes)
	hexString := hex.EncodeToString(randBytes)

	return hexString, nil
}
