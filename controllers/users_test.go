package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/stretchr/testify/assert"
)

func TestGetUsers(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	request := httptest.NewRequest(http.MethodGet, "http://users/", nil)
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	server.GetUsers(response, request, nil)
	assert.Equal(t, http.StatusOK, response.Code)
}
