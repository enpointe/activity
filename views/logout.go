package views

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// Logout log the user out deleting all cookies associated with session.
// The authentication JWT cookie, token, associated with the user session will be expired.
func (s *ServerService) Logout(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	token, httpStatus := validateClaim(response, request)
	if httpStatus != http.StatusOK {
		log.Infof("%s:%s successfully logged out, token expired", token.ID, token.Username)
		errorWithJSON(response, http.StatusText(httpStatus), httpStatus)
		return
	}

	http.SetCookie(response, &http.Cookie{
		Name:    "token",
		Value:   "",
		MaxAge:  -1, // Delete Now
		Expires: time.Now(),
	})

	log.Infof("%s:%s successfully logged out", token.ID, token.Username)
	response.WriteHeader(http.StatusOK)
}
