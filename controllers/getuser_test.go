package controllers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

// TestGetUserFailures test GetUser for some failure scenarios
func TestGetUserFailures(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	// Fail due to missing id parameter data
	request := httptest.NewRequest(http.MethodPost, "http:///activity/user/", nil)
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	server.GetUser(response, request, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, response.Code)

	// Attempt to retrieve details when token cookie is missing
	request = httptest.NewRequest(http.MethodGet, "http:///activity/user/"+testAdmin1ID, nil)
	response = httptest.NewRecorder()
	ps := httprouter.Params{
		httprouter.Param{
			Key:   "id",
			Value: testAdmin1ID},
	}
	server.GetUser(response, request, ps)
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
		t.Run(fmt.Sprintf("ID-%s", d.id),
			func(t *testing.T) {
				request := httptest.NewRequest(http.MethodGet, "http://users/"+d.id, nil)
				request.AddCookie(tokenCookie)
				response := httptest.NewRecorder()
				ps := httprouter.Params{
					httprouter.Param{
						Key:   "id",
						Value: d.id,
					},
				}
				server.GetUser(response, request, ps)
				assert.Equalf(t, d.expectedResponse, response.Code,
					"%s attempted to get user %s, expected '%s' got '%s'", creds.Username, d.id,
					http.StatusText(d.expectedResponse), http.StatusText(response.Code))
			})
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
		testGetUserData{"", http.StatusBadRequest},          // Attempt a get without specifying user
		testGetUserData{"doesNotExit", http.StatusNotFound}, // Attempt a get non-existent user
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
		testGetUserData{"", http.StatusBadRequest},          // Attempt a get without specifying user
		testGetUserData{"doesNotExit", http.StatusNotFound}, // Attempt a get non-existent user
	}
	getUserTest(t, creds, testData)
}

// TestGetUserViaBasicr test GetUser using the basic privilege user
func TestGetUserViaBasic(t *testing.T) {
	creds := client.Credentials{Username: testBasic1Username, Password: testBasic1UserPassword}
	testData := []testGetUserData{
		testGetUserData{testAdmin1ID, http.StatusForbidden},
		testGetUserData{testStaff1ID, http.StatusForbidden},
		testGetUserData{testStaff2ID, http.StatusForbidden},
		testGetUserData{testBasic1ID, http.StatusOK},
		testGetUserData{testBasic2ID, http.StatusForbidden},
		testGetUserData{"", http.StatusBadRequest},           // Attempt a get without specifying user
		testGetUserData{"doesNotExit", http.StatusForbidden}, // Attempt a get non-existent user
	}
	getUserTest(t, creds, testData)
}
