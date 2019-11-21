package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/perm"
	"github.com/stretchr/testify/assert"
)

func TestGetUsers(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	request := httptest.NewRequest("GET", "http://activity/users/", nil)
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	server.GetUsers(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Repeat as a post request
	request = httptest.NewRequest("POST", "http://activity/users/", nil)
	request.AddCookie(tokenCookie)
	response = httptest.NewRecorder()
	server.GetUsers(response, request)
	assert.Equal(t, http.StatusMethodNotAllowed, response.Code)
}

func TestCreateUser(t *testing.T) {
	server := setup(t, testAdminFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	u := client.UserCreate{
		Username:  "kristina",
		Password:  "changeMe",
		Privilege: perm.Staff.String(),
	}
	requestBody, err := json.Marshal(u)
	assert.NoError(t, err)
	request := httptest.NewRequest("POST", "http://user/create", bytes.NewBuffer(requestBody))
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	server.CreateUser(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Repeate as a GET request
	request = httptest.NewRequest("GET", "http://user/create", bytes.NewBuffer(requestBody))
	request.AddCookie(tokenCookie)
	response = httptest.NewRecorder()
	server.CreateUser(response, request)
	assert.Equal(t, http.StatusMethodNotAllowed, response.Code)

}

// Test to ensure basic privilege user can not perform delete action
func TestCreateBasicNoPrivileges(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testBasic1Username, Password: testBasic1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	u := client.UserCreate{
		Username:  "kristina",
		Password:  "changeMe",
		Privilege: perm.Staff.String(),
	}
	requestBody, err := json.Marshal(u)
	assert.NoError(t, err)
	request := httptest.NewRequest("POST", "http://user/create", bytes.NewBuffer(requestBody))
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	server.CreateUser(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

// Test to ensure staff privilege user can not create a admin user
func TestCreateAdminStaffPrivilege(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testStaff1Username, Password: testStaff1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	u := client.UserCreate{
		Username:  "kristina",
		Password:  "changeMe",
		Privilege: perm.Admin.String(),
	}
	requestBody, err := json.Marshal(u)
	assert.NoError(t, err)
	request := httptest.NewRequest("POST", "http://user/create", bytes.NewBuffer(requestBody))
	request.AddCookie(tokenCookie)
	response := httptest.NewRecorder()
	server.CreateUser(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}
