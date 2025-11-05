package auth

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nanagoboiler/models"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password []byte) (string, error) {

	passwordHash, err := bcrypt.GenerateFromPassword(password, 11)
	if err != nil {
		return "", err
	}

	return string(passwordHash), nil
}

func validateHashedPassword(rawPassword string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(rawPassword))

}

func validateCSRF(csrfCookie string, csrfHeader string) error {
	if csrfCookie == "" || csrfHeader == "" {
		return fmt.Errorf("csrf validation failed")
	}

	if csrfCookie != csrfHeader {
		return fmt.Errorf("csrf validation failed")
	}
	return nil
}

func validateJWT(token string) (models.User, error) {
	secret := os.Getenv("JWT_SECRET")

	jwtToken, err := jwt.ParseWithClaims(token, &models.AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return models.User{}, err
	}

	if claims, ok := jwtToken.Claims.(*models.AuthClaims); ok && jwtToken.Valid {

		return models.User{
			ID:       claims.UserId,
			Username: claims.UserName,
		}, nil
	}

	return models.User{}, fmt.Errorf("invalid token")

}
