package views

import (
	"net/http"
	"time"
)

// Logout log the user out deleting all cookies associated with session
func Logout(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("content-type", "application/json")
	_, httpStatus := validateClaim(response, request)
	if httpStatus != http.StatusOK {
		response.WriteHeader(httpStatus)
		return
	}

	http.SetCookie(response, &http.Cookie{
		Name:    "token",
		Value:   "",
		MaxAge:  -1, // Delete Now
		Expires: time.Now(),
	})
	response.WriteHeader(http.StatusOK)
}
