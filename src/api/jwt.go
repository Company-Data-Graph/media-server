package api

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"media-server/src/models"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var users map[string]string

var jwtKey []byte

func GenerateNewToken(creds models.Credentials, t int) (string, error) {
	if len(jwtKey) == 0 {
		log.Println("JwtKey not found! Generate new key!")
		jwtKey = []byte(base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", rand.Intn(1000000)))))
		log.Println("New key:", string(jwtKey))
	}
	expirationTime := time.Now().Add(time.Minute * time.Duration(t))
	claims := &models.Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	return tokenString, err
}

func CheckToken() {

}

func CredentialsValidation(creds models.Credentials) bool {
	expectedCreds, ok := users[creds.Username]
	if !ok || creds.Password != expectedCreds {
		return false
	}
	return true
}
