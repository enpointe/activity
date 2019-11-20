package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/stretchr/testify/assert"
)

// TestGetUserFailures test GetUser for some failure scenarios
func TestGetUserFailures(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	// Fail due to POST
	request := httptest.NewRequest("POST", "http:///activity/user/", nil)
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusMethodNotAllowed, response.Code)

	// Attempt to retrieve details when token cookie is missing
	request = httptest.NewRequest("GET", "http:///activity/user/"+testAdmin1ID, nil)
	response = httptest.NewRecorder()
	server.GetUser(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

type testGetUserData struct {
	id               string
	expectedResponse int
}

func getUserTest(t *testing.T, creds client.Credentials, testData []testGetUserData) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	for _, d := range testData {
		request := httptest.NewRequest("GET", "http:///activity/user/"+d.id, nil)
		request.AddCookie(tokenCookie)
		response := httptest.NewRecorder()
		server.GetUser(response, request)
		assert.Equalf(t, d.expectedResponse, response.Code,
			"%s attempted to get user %s, expected '%s' got '%s'", creds.Username, d.id,
			http.StatusText(d.expectedResponse), http.StatusText(response.Code))
	}
}

// TestGetUserViaAdmin test GetUser using a user with perm.ADMIN privileges
func TestGetUserViaAdmin(t *testing.T) {
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	testData := []testGetUserData{
		testGetUserData{testAdmin1ID, http.StatusOK},
		testGetUserData{testAdmin2ID, http.StatusOK},
		testGetUserData{testStaff1ID, http.StatusOK},
		testGetUserData{testBasic1ID, http.StatusOK},
		testGetUserData{"", http.StatusBadRequest},            // Attempt a get without specifying user
		testGetUserData{"doesNotExit", http.StatusBadRequest}, // Attempt a get non-existent user
	}
	getUserTest(t, creds, testData)
}

// TestGetUserViaStaff test GetUser using a user with perm.STAFF privileges
func TestGetUserViaStaff(t *testing.T) {
	creds := client.Credentials{Username: testStaff1Username, Password: testStaff1UserPassword}
	testData := []testGetUserData{
		testGetUserData{testAdmin1ID, http.StatusOK},
		testGetUserData{testStaff1ID, http.StatusOK},
		testGetUserData{testStaff2ID, http.StatusOK},
		testGetUserData{testBasic1ID, http.StatusOK},
		testGetUserData{"", http.StatusBadRequest},            // Attempt a get without specifying user
		testGetUserData{"doesNotExit", http.StatusBadRequest}, // Attempt a get non-existent user
	}
	getUserTest(t, creds, testData)
}

// TestGetUserViaBasicr test GetUser using the basic privilege user
func TestGetUserViaBasic(t *testing.T) {
	creds := client.Credentials{Username: testBasic1Username, Password: testBasic1UserPassword}
	testData := []testGetUserData{
		testGetUserData{testAdmin1ID, http.StatusUnauthorized},
		testGetUserData{testStaff1ID, http.StatusUnauthorized},
		testGetUserData{testStaff2ID, http.StatusUnauthorized},
		testGetUserData{testBasic1ID, http.StatusOK},
		testGetUserData{testBasic2ID, http.StatusUnauthorized},
		testGetUserData{"", http.StatusBadRequest},              // Attempt a get without specifying user
		testGetUserData{"doesNotExit", http.StatusUnauthorized}, // Attempt a get non-existent user
	}
	getUserTest(t, creds, testData)
}
