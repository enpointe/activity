package controllers

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// Logout log the user out deleting all cookies associated with session.
// By nature of how JWT tokens work if the token is cached, the old
// token can continue to be recognized as authorized till the token
// expires.
//
// @Summary Log out the current user
// @Description Log out the current user.
// @Description By nature of how JWT tokens work if the token is cached, the old
// @Description token can continue to be recognized as authorized till the token
// @Description expires
//
// @Security ApiKeyAuth
// @in header
// @name Authorization
// @param Authorization header string true "The JWT authorization token acquired at login"
// @Accept  json
// @Produce  json
// @Success 200 {object} string "OK"
// @Failure 400 {object} APIError "Bad Request"
// @Failure 401 {object} APIError "Unauthorized"
// @Failure 500 {object} APIError "Internal Server Error"
// @Router /logout [post]
func (s *ServerService) Logout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	token, httpStatus := validateClaim(w, r)
	if httpStatus != http.StatusOK {
		if token != nil {
			log.Infof("%s:%s successfully logged out, token expired", token.ID, token.Username)
		}
		errorWithJSON(w, http.StatusText(httpStatus), httpStatus)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   "",
		MaxAge:  -1, // Delete Now
		Expires: time.Now(),
	})

	log.Infof("%s:%s successfully logged out", token.ID, token.Username)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
}
