package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/damascopaul/lfg-backend/endpoints"
	"github.com/damascopaul/lfg-backend/schemas"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

func parseJwt(t string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error("Could not parse JWT. Unexpected signing method")
			return nil, fmt.Errorf(
				"unexpected signing method. Method: %v", token.Header)
		}
		return []byte(endpoints.TOKEN_SECRET), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	log.Errorf("Could not parse JWT. Error: %v", err)
	return jwt.MapClaims{}, err
}

// AuthenticateRequests checks if the request is authorized.
//
// This checks the JWT in the `Authorization` header.
func AuthenticateRequests(c *gin.Context) {
	// TODO: Add checking of iat value.
	ah := c.Request.Header.Get("Authorization")
	if ah == "" {
		log.Error("Could not authenticate request. Authorization header is missing")
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			schemas.BodyError{Message: "Authorization header is missing"})
		return
	}
	token := strings.Split(ah, " ")[1]
	claims, err := parseJwt(token)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected signing method") {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				schemas.BodyError{Message: "Token is invalid"})
			return
		} else {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError, endpoints.BodyInternalServerError)
			return
		}
	}
	uid := claims["user_id"].(float64)
	c.Set("user_id", int64(uid))
	c.Next()
}
