package views_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/enpointe/activity/views"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const testDatabase string = "Activity_HTTP_Test"
const testAdminFilenameJSON string = "testdata/admin_user.json"
const testMultiUserFilenameJSON string = "testdata/multiuser_data.json"

// The usernames and the passwords here correspond
// to the usernames added from the testdata json file
const testAdmin1Username string = "admin1"
const testAdmin1UserPassword string = "changeMe"
const testStaff1Username string = "staff1"
const testStaff1UserPassword string = testAdmin1UserPassword
const testBasic1Username string = "customer1"
const testBasic1UserPassword string = testAdmin1UserPassword
const testAdmin2Username string = "admin2"
const testAdmin2UserPassword string = "changeMe"
const testStaff2Username string = "staff2"
const testStaff2UserPassword string = testAdmin1UserPassword
const testBasic2Username string = "customer2"
const testBasic2UserPassword string = testAdmin1UserPassword

// setup Setup the database for testing by creating a connection to the
// database and returning a handle to the UserService. If desired
// via the clear flag the current user collection entires can be
// dropped. Setting the load flag causes the predefined user collection
// entires in TestUserFilename to be inserted into the user collection.
func setup(t *testing.T, userLoadFile string) *views.ServerService {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	opt := views.DBOptions(clientOptions)
	logger := log.StandardLogger()

	// We need to have at least one admin user present in our database
	// to proceed.

	server, err := views.NewServerService(true, opt,
		views.DBName(testDatabase), views.Log(logger))
	assert.NoError(t, err)
	err = server.DeleteAll()
	assert.NoError(t, err)
	uService, err := db.NewUserService(server.Database, logger)
	assert.NoError(t, err)
	err = uService.LoadFromFile(context.TODO(), userLoadFile)
	assert.NoErrorf(t, err, "Error loading file %s", userLoadFile)
	return server
}

// teardown - perform database teardown to ensure each
// that the database is clean
func teardown(t *testing.T, server *views.ServerService) {
	err := server.DeleteAll()
	assert.NoError(t, err)
}

// login helper function that logs the specified user in and
// returns the JWT authentication token to use on subsequent request
func login(t *testing.T, server *views.ServerService, creds client.Credentials) *http.Cookie {
	requestBody, err := json.Marshal(map[string]string{
		"username": creds.Username,
		"password": creds.Password,
	})
	assert.NoError(t, err)
	request := httptest.NewRequest("POST", "http://login", bytes.NewBuffer(requestBody))
	response := httptest.NewRecorder()
	server.Login(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Check to make sure the cookie token is present
	cookies := response.Result().Cookies()
	var tokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "token" {
			tokenCookie = c
		}
	}
	assert.NotNil(t, tokenCookie)
	return tokenCookie
}

// logout logs the user out
func logout(t *testing.T, server *views.ServerService, tokenCookie *http.Cookie) {
	request := httptest.NewRequest("POST", "http://logout", nil)
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	server.Logout(response, request)
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
	request := httptest.NewRequest("POST", "http://login", bytes.NewBuffer(requestBody))
	response := httptest.NewRecorder()
	server.Login(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	// Test bad password
	requestBody, err = json.Marshal(map[string]string{
		"username": testAdmin1Username,
		"password": "badPassword",
	})
	assert.NoError(t, err)
	request = httptest.NewRequest("POST", "http://login", bytes.NewBuffer(requestBody))
	response = httptest.NewRecorder()
	server.Login(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	// Test GET request instead of a POST
	requestBody, err = json.Marshal(map[string]string{
		"username": testAdmin1Username,
		"password": "badPassword",
	})
	assert.NoError(t, err)
	request = httptest.NewRequest("GET", "http://login", bytes.NewBuffer(requestBody))
	response = httptest.NewRecorder()
	server.Login(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}
