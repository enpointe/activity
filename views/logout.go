package views

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// Logout log the user out deleting all cookies associated with session
func (s *ServerService) Logout(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	token, httpStatus := validateClaim(response, request)
	if httpStatus != http.StatusOK {
		response.WriteHeader(httpStatus)
		log.Infof("%s:%s successfully logged out, token expired", token.ID, token.Username)
		return
	}

	http.SetCookie(response, &http.Cookie{
		Name:    "token",
		Value:   "",
		MaxAge:  -1, // Delete Now
		Expires: time.Now(),
	})

	log.Infof("successfully logged out %s:%s", token.ID, token.Username)
	response.WriteHeader(http.StatusOK)
}
