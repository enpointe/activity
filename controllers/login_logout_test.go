package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enpointe/activity/controllers"
	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const testDatabase string = "Activity_HTTP_Test"
const testAdminFilenameJSON string = "testdata/admin_user.json"
const testMultiUserFilenameJSON string = "testdata/multiuser_data.json"

// The usernames and the passwords here correspond
// to the usernames added from the testdata json file
const testAdmin1ID string = "5db8e02b0e7aa732afd7fbc3"
const testAdmin1Username string = "admin1"
const testAdmin1UserPassword string = "changeMe"
const testStaff1ID string = "5db8e02b0e7aa732afd7fbc2"
const testStaff1Username string = "staff1"
const testStaff1UserPassword string = testAdmin1UserPassword
const testBasic1ID string = "5db8e02b0e7aa732afd7fbc1"
const testBasic1Username string = "customer1"
const testBasic1UserPassword string = testAdmin1UserPassword
const testAdmin2ID string = "5db8e02b0e7aa732afd7fbc6"
const testAdmin2Username string = "admin2"
const testAdmin2UserPassword string = "changeMe"
const testStaff2ID string = "5db8e02b0e7aa732afd7fbc5"
const testStaff2Username string = "staff2"
const testStaff2UserPassword string = testAdmin1UserPassword
const testBasic2ID string = "5db8e02b0e7aa732afd7fbc4"
const testBasic2Username string = "customer2"
const testBasic2UserPassword string = testAdmin1UserPassword

// setup Setup the database for testing by creating a connection to the
// database and returning a handle to the UserService. If desired
// via the clear flag the current user collection entires can be
// dropped. Setting the load flag causes the predefined user collection
// entires in TestUserFilename to be inserted into the user collection.
func setup(t *testing.T, userLoadFile string) *controllers.ServerService {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	opt := controllers.DBOptions(clientOptions)

	// We need to have at least one admin user present in our database
	// to proceed.

	server, err := controllers.NewServerService(true, opt,
		controllers.DBName(testDatabase))
	assert.NoError(t, err)
	err = server.DeleteAll()
	assert.NoError(t, err)
	uService, err := db.NewUserService(server.Database)
	assert.NoError(t, err)
	if len(userLoadFile) > 0 {
		err = uService.LoadFromFile(context.TODO(), userLoadFile)
		assert.NoErrorf(t, err, "Error loading file %s", userLoadFile)
	}
	return server
}

// teardown - perform database teardown to ensure each
// that the database is clean
func teardown(t *testing.T, server *controllers.ServerService) {
	err := server.DeleteAll()
	assert.NoError(t, err)
}

// login helper function that logs the specified user in and
// returns the JWT authentication token to use on subsequent request
func login(t *testing.T, server *controllers.ServerService, creds client.Credentials) *http.Cookie {
	requestBody, err := json.Marshal(creds)
	assert.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "http://login", bytes.NewBuffer(requestBody))
	response := httptest.NewRecorder()
	server.Login(response, request, nil)
	assert.Equal(t, http.StatusOK, response.Code)

	// Check to make sure the cookie token is present
	cookies := response.Result().Cookies()
	var tokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == controllers.TokenCookie {
			tokenCookie = c
		}
	}
	assert.NotNil(t, tokenCookie)
	return tokenCookie
}

// logout logs the user out
func logout(t *testing.T, server *controllers.ServerService, tokenCookie *http.Cookie) {
	request := httptest.NewRequest(http.MethodPost, "http://logout", nil)
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	server.Logout(response, request, nil)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestLoginLogout(t *testing.T) {
	server := setup(t, testAdminFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	tokenCookie := login(t, server, creds)
	logout(t, server, tokenCookie)
}

func TestInvalidLogin(t *testing.T) {
	server := setup(t, testAdminFilenameJSON)
	defer teardown(t, server)

	// Test bad username
	requestBody, err := json.Marshal(map[string]string{
		"username": "badUser",
		"password": "",
	})
	assert.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "http://login", bytes.NewBuffer(requestBody))
	response := httptest.NewRecorder()
	server.Login(response, request, nil)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	// Test bad password
	requestBody, err = json.Marshal(map[string]string{
		"username": testAdmin1Username,
		"password": "badPassword",
	})
	assert.NoError(t, err)
	request = httptest.NewRequest(http.MethodPost, "http://login", bytes.NewBuffer(requestBody))
	response = httptest.NewRecorder()
	server.Login(response, request, nil)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	// Test GET request instead of a POST
	requestBody, err = json.Marshal(map[string]string{
		"username": testAdmin1Username,
		"password": "badPassword",
	})
	assert.NoError(t, err)
	request = httptest.NewRequest(http.MethodGet, "http://login", bytes.NewBuffer(requestBody))
	response = httptest.NewRecorder()
	server.Login(response, request, nil)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestLogoutNoToken(t *testing.T) {
	server := setup(t, testAdminFilenameJSON)
	defer teardown(t, server)
	request := httptest.NewRequest(http.MethodPost, "http://logout", nil)
	response := httptest.NewRecorder()
	server.Logout(response, request, nil)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestReuseOfToken(t *testing.T) {
	server := setup(t, testAdminFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	tokenCookie := login(t, server, creds)
	copy := *tokenCookie
	logout(t, server, &copy)

	// Try to perform an action using the original token cookie
	// This should fail because the user logged out.
	request := httptest.NewRequest(http.MethodGet, "http:///activity/user/"+testBasic1ID, nil)
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	ps := httprouter.Params{
		httprouter.Param{
			Key:   "id",
			Value: testBasic1ID,
		},
	}
	server.GetUser(response, request, ps)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}
