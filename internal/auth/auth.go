package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("couldn't hash password: %w", err)
    }
    return string(hashedPassword), nil
}

func CheckPasswordHash(password, hashPassword string) error {
    err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
    return err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
    claims := jwt.RegisteredClaims{
        Issuer:    "chirpy",
        Subject:   userID.String(),
        IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
        ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
    }

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    signedToken, err := token.SignedString([]byte(tokenSecret))
    if err != nil {
        return "", fmt.Errorf("failed to sign token: %w", err)
    }

	fmt.Printf("Signed Token: %s\n", signedToken)

    return signedToken, nil
}

// ValidateJWT parses and validates the token, returning the user ID if valid.
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
    token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
        // Provide the same secret used in MakeJWT
        return []byte(tokenSecret), nil
    })

    if err != nil {
        return uuid.Nil, fmt.Errorf("invalid token: %w", err)
    }

    // Assert that the claims in the token are of type *jwt.RegisteredClaims
    claims, ok := token.Claims.(*jwt.RegisteredClaims)
    if !ok || !token.Valid {
        return uuid.Nil, fmt.Errorf("invalid token claims")
    }

    // Convert the Subject field (string) back to a UUID
    userID, err := uuid.Parse(claims.Subject)
    if err != nil {
        return uuid.Nil, fmt.Errorf("invalid user id in token subject: %w", err)
    }

    return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
    authHeader := headers.Get("Authorization")
    if authHeader == "" {
        return "", errors.New("no Authorization header")
   }

   	authParts := strings.Split(authHeader, " ")

	if len(authParts) != 2 {
		return "", errors.New("authorization header format must be 'Bearer {token}")
	}

	authType := authParts[0]
	authToken := authParts[1]

    if authType != "Bearer" {
        return "", errors.New("invalid authorization type")
    }

    return strings.TrimSpace(authToken), nil
}

func MakeRefreshToken() (string, error) {
    key := make([]byte, 32)

    _, err := rand.Read(key)
    if err != nil {
        return "", errors.New("Unable to read key data")
    }

    hexString := hex.EncodeToString(key)

    return hexString, nil
}

func GetAPIKey(headers http.Header) (string, error) {
    authHeader := headers.Get("Authorization")
    if authHeader == "" {
        return "", errors.New("no Authorization header")
   }

   	authParts := strings.Split(authHeader, " ")

	if len(authParts) != 2 {
		return "", errors.New("authorization header format must be 'APIKEY {apikey}")
	}

    if authParts[0] != "ApiKey" {
        return "", errors.New("authorization header format must be 'APIKEY {apikey}")
    }
    apiKey := authParts[1]

    return strings.TrimSpace(apiKey), nil
}
