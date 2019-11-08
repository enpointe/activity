package views_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/stretchr/testify/assert"
)

// TestGetUserFailures test GetUser for general non permission failure scenarios
func TestGetUserFailures(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	// Attempt to retrieve user details without specifying user
	request := httptest.NewRequest("GET", "http:///activity/user/", nil)
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// Attempt to retrieve details about non-exist user
	request = httptest.NewRequest("GET", "http:///activity/user/doesNotExist", nil)
	request.AddCookie(tokenCookie)
	response = httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// Attempt to retrieve details when token has expired
	request = httptest.NewRequest("GET", "http:///activity/user/doesNotExist", nil)
	tokenCookie.MaxAge = 0
	request.AddCookie(tokenCookie)
	response = httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

// TestGetUserViaAdmin test GetUser using a user with perm.ADMIN privileges
func TestGetUserViaAdmin(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	// Attempt to retrieve details about the same user
	request := httptest.NewRequest("GET", "http:///activity/user/"+testAdmin1ID, nil)
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Attempt to retrieve details about the another admin user
	request = httptest.NewRequest("GET", "http:///activity/user/"+testAdmin2ID, nil)
	request.AddCookie(tokenCookie)
	response = httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Attempt to retrieve details about the admin user
	request = httptest.NewRequest("GET", "http:///activity/user/"+testStaff1ID, nil)
	request.AddCookie(tokenCookie)
	response = httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Attempt to retrieve details about the customer user
	request = httptest.NewRequest("GET", "http:///activity/user/"+testBasic1ID, nil)
	request.AddCookie(tokenCookie)
	response = httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

}

// TestGetUserViaStaff test GetUser using a user with perm.STAFF privileges
func TestGetUserViaStaff(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testStaff1Username, Password: testStaff1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	// Staff user can retrieve a user with perm.ADMIN privilege,
	// as no password information is sent back
	request := httptest.NewRequest("GET", "http:///activity/user/"+testAdmin1ID, nil)
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Staff user can retrieve a user with perm.STAFF privilege
	request = httptest.NewRequest("GET", "http:///activity/user/"+testStaff1ID, nil)
	request.AddCookie(tokenCookie)
	response = httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Staff user can retrieve a user with perm.BASIC privilege
	request = httptest.NewRequest("GET", "http:///activity/user/"+testBasic1ID, nil)
	request.AddCookie(tokenCookie)
	response = httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
}

// TestGetUserViaBasicr test GetUser using the basic privilege user
func TestGetUserViaBasic(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testBasic1Username, Password: testBasic1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	// Verify perm.BASIC will not allow retrieval of perm.ADMIN privilege user
	request := httptest.NewRequest("GET", "http:///activity/user/"+testAdmin1ID, nil)
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	// Verify perm.BASIC will not allow retrieval of perm.STAFF privilege user
	request = httptest.NewRequest("GET", "http:///activity/user/"+testStaff1ID, nil)
	request.AddCookie(tokenCookie)
	response = httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	// Staff user can retrieve a user with perm.BASIC privilege
	request = httptest.NewRequest("GET", "http:///activity/user/"+testBasic1ID, nil)
	request.AddCookie(tokenCookie)
	response = httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
}
