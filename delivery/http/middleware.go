package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)

type JWTClaim struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Type      string `json:"type"`
	IsPremium bool   `json:"is_premium"`
	jwt.StandardClaims
}

func Auth() gin.HandlerFunc {
	return func(context *gin.Context) {
		authorization := context.GetHeader("Authorization")
		if authorization == "" {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "request does not contain an access token"})
			context.Abort()
			return
		}
		_, tokenString, _ := strings.Cut(authorization, " ")
		claims, err := validateToken(tokenString)
		if err != nil {
			context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			context.Abort()
			return
		}

		context.Set("username", claims.Username)
		context.Set("type", claims.Type)
		context.Set("is_premium", claims.IsPremium)
		context.Next()
	}
}

func Admin() gin.HandlerFunc {
	return func(context *gin.Context) {
		typeAccount := context.GetString("type")
		if typeAccount != "ADMIN" {
			context.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unauthorized"})
			context.Abort()
			return
		}
		context.Next()
	}
}

func Premium() gin.HandlerFunc {
	return func(context *gin.Context) {
		isPremium := context.GetBool("is_premium")
		if !isPremium {
			context.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unauthorized"})
			context.Abort()
			return
		}
		context.Next()
	}
}

func validateToken(signedToken string) (claims *JWTClaim, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(viper.GetString("jwt_secret")), nil
		},
	)
	if err != nil {
		return
	}
	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return
	}
	return
}
