package services

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"main/config"
	"net/http"
	"strings"
	"time"
)

func AuthenticateUser(context *gin.Context) {

	tokenString := GetTokenFromContext(context)
	if tokenString == "" {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "token is empty or malformed"})
		return
	}

	token, err := ParseToken(tokenString)
	if err != nil || token == nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to get subject"})
		return
	}

	if subject == "" {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Check if the token is still valid
	expiration, err := token.Claims.GetExpirationTime()
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token expiry"})
		return
	}
	if expiration.Before(time.Now()) {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is expired"})
		return
	}

}

func GenerateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(time.Duration(config.Settings.MaxTokenLifeMinutes) * time.Minute)

	// Create the claims
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		Subject:   username,
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(config.Settings.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetTokenFromContext(context *gin.Context) string {
	// Check if token exists on the request
	tokenString := context.GetHeader("Authorization")
	if tokenString == "" {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
		return ""
	}

	// Remove the headers prefix 'Bearer' so that it can be parsed
	split := strings.Split(tokenString, " ")

	if len(split) != 2 {
		logrus.Error("Invalid Authorization header")
		return ""
	}

	if split[0] != "Bearer" {
		logrus.Error("Missing Bearer in auth header")
		return ""
	}

	if split[1] == "" {
		logrus.Error("Missing jwt string")
		return ""
	}

	return split[1]
}

func GetSubjectFromContext(context *gin.Context) string {
	tokenString := GetTokenFromContext(context)
	if tokenString == "" {
		logrus.Error("Cannot get subject from token")
		return ""
	}

	parsedToken, err := ParseToken(tokenString)
	if err != nil || parsedToken == nil {
		logrus.Error("Cannot parse token")
		return ""
	}

	subject, err := parsedToken.Claims.GetSubject()
	if err != nil || subject == "" {
		logrus.Error("Cannot get subject from token")
		return ""
	}

	return subject
}

func ParseToken(tokenString string) (*jwt.Token, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(config.Settings.SecretKey), nil
	})

	return token, err
}
