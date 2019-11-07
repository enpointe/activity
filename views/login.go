package views

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

// Login interface for allowing the user to acquire authorization
// to execute methods for this application
func (s *ServerService) Login(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	if request.Method != "POST" {
		log.Warning("unauthorized GET login attempt")
		response.WriteHeader(http.StatusUnauthorized)
		return
	}

	var creds client.Credentials
	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil {
		log.Warningf("invalid log attempt, bad payload: %s", request.Body)
		response.WriteHeader(http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
	defer cancel()

	userService, err := db.NewUserService(s.client, s.Database)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	clientUser, err := userService.Validate(ctx, &creds)
	if err != nil {
		log.Warning("Credentials didn't validate")
		response.WriteHeader(http.StatusUnauthorized)
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
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Finally, we set the client cookie for "token" as the JWT we just generated
	// we also set an expiry time which is the same as the token itself
	http.SetCookie(response, &http.Cookie{
		Name:   "token",
		Value:  tokenString,
		MaxAge: jwtExpirySeconds * 1000,
	})
	log.Infof("successfully logged in %s:%s", clientUser.ID, clientUser.Username)
	response.WriteHeader(http.StatusOK)
}
