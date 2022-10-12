package endpoints

import (
	"net/http"
	"strings"
	"time"

	"github.com/damascopaul/lfg-backend/schemas"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func buildResponseWithToken(u schemas.User) (schemas.TokenResponse, error) {
	claim := createJWTClaim(u)
	jwt, err := generateJWT(claim, []byte(TOKEN_SECRET))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorf("Could not build response body")
		return schemas.TokenResponse{}, err
	}

	u.Password = "" // Removes the password from the response
	r := schemas.TokenResponse{
		Token: jwt,
		User:  u,
	}
	log.Info("Response body built")
	return r, nil
}

func createJWTClaim(u schemas.User) jwt.MapClaims {
	c := jwt.MapClaims{
		"user_id":  u.ID,
		"username": u.Username,
		"iat":      time.Now().Unix(),
	}
	return c
}

func generateJWT(c jwt.MapClaims, secret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	jwt, err := token.SignedString(secret)
	if err != nil {
		log.Errorf("Could not generate JWT. Error: %v", err)
		return "", err
	}
	return jwt, nil
}

// SignUp allows users to create an account.
func SignUp(c *gin.Context) {
	u, _ := c.Keys["req"].(schemas.User)

	if err := u.ValidateForSignUp(); err != nil {
		log.WithFields(log.Fields{
			"endpoint": "SignUp",
			"error":    err.Error(),
		}).Warn("Request failed")
		validationError, _ := err.(*schemas.ValidationError)
		c.AbortWithStatusJSON(http.StatusBadRequest, schemas.BodyError{
			Message:     err.Error(),
			FieldErrors: validationError.Errors,
		})
		return
	}

	if err := u.InitDB(); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	err := u.Create()
	if err != nil {
		const usernameError = "UNIQUE constraint failed: users.username"
		if err.Error() == usernameError {
			// Return a 404 error if the error is related to
			// the uniqueness of the username.
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				schemas.BodyError{Message: "User already exists."})
			return
		}
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	resp, err := buildResponseWithToken(u)
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}
	c.JSON(http.StatusCreated, resp)
	log.WithFields(log.Fields{"endpoint": "SignUp"}).Info("Request successful")
}

// SignIn allows existing users to sign in with their username and password.
func SignIn(c *gin.Context) {
	u, _ := c.Keys["req"].(schemas.User)
	reqPW := u.Password

	bodyInvalidCredentials := schemas.BodyError{
		Message: "username or password is invalid.",
	}

	if err := u.InitDB(); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	err := u.RetrieveByUsername()
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			// Return a 403 error if there is
			// no matching user given the username
			c.AbortWithStatusJSON(
				http.StatusUnauthorized, bodyInvalidCredentials)
			return
		}
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		// Return a 403 error if the password does not match
		[]byte(u.Password), []byte(reqPW)); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, bodyInvalidCredentials)
		return
	}

	resp, err := buildResponseWithToken(u)
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}
	c.JSON(http.StatusCreated, resp)
	log.WithFields(log.Fields{"endpoint": "SignIn"}).Info("Request successful")
}
