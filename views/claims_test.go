package views_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/stretchr/testify/assert"
)

// TestValidateClaims
func TestMissingClaimsCookie(t *testing.T) {
	server := setup(t, "")
	request := httptest.NewRequest("GET", "http://users", nil)
	response := httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestInvalidClaim(t *testing.T) {
	server := setup(t, testAdminFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	tokenCookie := login(t, server, creds)

	// Attempt to retrieve with an expired JWT token
	tokenToModify := *tokenCookie
	tokenToModify.MaxAge = 0
	request := httptest.NewRequest("GET", "http:///activity/users/", nil)
	request.AddCookie(&tokenToModify)
	response := httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}
