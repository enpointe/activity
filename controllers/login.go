package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/enpointe/activity/perm"
	log "github.com/sirupsen/logrus"
)

const jwtExpirySeconds = 1200

// TODO - allow jwtKey to be set via CLI interface
var jwtKey = []byte("my_secret_key")

type activityClaims struct {
	ID        string
	Username  string
	Privilege perm.Privilege
	jwt.StandardClaims
}

// Login interface for allowing the user to acquire authorization to execute methods
// for this application. The privileges associated with a users account (client.UserInfo.Privilege)
// will dictat what methods can be invoked by the user.
// @Summary Login log a user into server
// @Description Log a user into the activity server, allowing the user to
// @Description acquire authorization to execute methods for this application.
// @Description The privileges associated with a users account (client.UserInfo.Privilege)
// @Description will dictat what methods can be invoked by the user.
// @Tags client.Credentials, client.UserInfo
// @Param Credentials body client.Credentials true "Login Credentials"
// @Accept  json
// @Produce  json
// @Success 200 {object} client.UserInfo
// @Header 200 {string} auth "JWT Authentication Token"
// @Failure 400 {object} APIError "Bad Request"
// @Failure 401 {object} APIError "Unauthorized"
// @Failure 404 {object} APIError "Not Found"
// @Failure 405 {object} APIError "Method Not Allowed"
// @Failure 500 {object} APIError "Internal Server Error"
// @Security ApiKeyAuth
// @Router /login [post]
func (s *ServerService) Login(response http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		errorWithJSON(response,
			http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var creds client.Credentials
	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil {
		log.Warningf("invalid log attempt, bad payload: %s", request.Body)
		errorWithJSON(response,
			http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
	defer cancel()

	userService, err := db.NewUserService(s.Database)
	if err != nil {
		errorWithJSON(response,
			http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	clientUser, err := userService.Validate(ctx, &creds)
	if err != nil {
		log.Warning("Credentials didn't validate")
		errorWithJSON(response,
			http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Declare the expiration time of the token
	// here, we have kept it as expireTime minutes
	expirationTime := time.Now().Add(time.Duration(jwtExpirySeconds) * time.Second)
	// Create the JWT claims, which includes the username and expiry time
	claims := &activityClaims{
		ID:        clientUser.ID,
		Username:  clientUser.Username,
		Privilege: perm.Convert(clientUser.Privilege),
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string.
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		log.Errorf("JWT signing issue: %s", err)
		errorWithJSON(response,
			http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Finally, we set the client cookie for "token" as the JWT we just generated
	// we also set an expiry time which is the same as the token itself
	http.SetCookie(response, &http.Cookie{
		Name:   TokenCookie,
		Value:  tokenString,
		MaxAge: jwtExpirySeconds * 1000,
	})
	log.Infof("successfully logged in %s:%s", clientUser.ID, clientUser.Username)
	response.Header().Set("content-type", "application/json")
	json.NewEncoder(response).Encode(clientUser)
	response.WriteHeader(http.StatusOK)
}
