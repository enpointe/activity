package views

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/enpointe/activity/perm"
)

// Claims the JWS Claims structure used to authenticate
// a user and privileges once logged in. ID represents
// the identifier for the user. Username repesents
// the login name of the user. Privilege represents
// the privilege level the user pocesses in order
// to perform actions against the http JSON interface.
type Claims struct {
	ID        string         `json:"id"`
	Username  string         `json:"username"`
	Privilege perm.Privilege `json:"privilege"`
	jwt.StandardClaims
}

// validateClaim validate the JWT token string stored in the token cookie
// Returns the claims structure if the JWT claim is validated. Returns http error
// status code if the claim fails.
func validateClaim(response http.ResponseWriter, request *http.Request) (*Claims, int) {
	c, err := request.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			return nil, http.StatusUnauthorized
		}
		// For any other type of error, return a bad request status
		return nil, http.StatusBadRequest
	}

	// Get the JWT string from the cookie
	tknStr := c.Value

	// Initialize a new instance of `Claims`
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, http.StatusUnauthorized
		}
		return nil, http.StatusBadRequest
	}
	if !tkn.Valid {
		return nil, http.StatusUnauthorized
	}
	return claims, http.StatusOK
}
