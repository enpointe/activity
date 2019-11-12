package views_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/stretchr/testify/assert"
)

type uPasswdTestData struct {
	userInfo         client.User
	expectedResponse int
}

func TestUpdateFailures(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testBasic1Username, Password: testBasic1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	type testData struct {
		method           string
		testUser         string
		expectedResponse int
	}

	testInput := []testData{
		testData{ // Incorrect Method
			method:           "GET",
			testUser:         testBasic1ID,
			expectedResponse: http.StatusMethodNotAllowed,
		},
		testData{ // No user specified
			method:           "POST",
			testUser:         "",
			expectedResponse: http.StatusBadRequest,
		},
		testData{ // Request is fine but no JSON data
			method:           "POST",
			testUser:         testBasic1ID,
			expectedResponse: http.StatusBadRequest,
		},
	}
	for _, data := range testInput {
		url := "http://user/Update/" + data.testUser
		request := httptest.NewRequest(data.method, url, nil)
		request.AddCookie(tokenCookie)
		response := httptest.NewRecorder()
		server.UpdateUserPassword(response, request)
		assert.Equalf(t, data.expectedResponse, response.Code, "%s %s", data.method, url)
	}

	// Test no token associated with request
	request := httptest.NewRequest("POST", "http://user/Update/"+testBasic1ID, nil)
	response := httptest.NewRecorder()
	server.UpdateUserPassword(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestAdminUpdatePassword(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	// Attempt to change the password of various users
	newPassword := "newPasswordValue"
	testUsers := []client.User{
		client.User{
			Username: testAdmin1Username,
			Password: newPassword,
		},
		client.User{
			Username: testStaff1Username,
			Password: newPassword,
		},
		client.User{
			Username: testBasic1Username,
			Password: newPassword,
		},
	}
	for _, user := range testUsers {
		requestBody, err := json.Marshal(user)
		assert.NoError(t, err)
		request := httptest.NewRequest("POST", "http://user/UpdateUserPassword", bytes.NewBuffer(requestBody))
		request.AddCookie(tokenCookie)
		response := httptest.NewRecorder()
		server.UpdateUserPassword(response, request)
		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestStaffUpdatePassword(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testStaff1Username, Password: testStaff1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	// Attempt to change the password of various users
	newPassword := "newPasswordValue"
	testData := []uPasswdTestData{
		uPasswdTestData{
			userInfo: client.User{
				Username: testAdmin1Username,
				Password: newPassword,
			},
			expectedResponse: http.StatusUnauthorized,
		},
		uPasswdTestData{
			userInfo: client.User{
				Username: testStaff1Username,
				Password: newPassword,
			},
			expectedResponse: http.StatusOK,
		},

		uPasswdTestData{
			userInfo: client.User{
				Username: testStaff2Username,
				Password: newPassword,
			},
			expectedResponse: http.StatusUnauthorized,
		},
		uPasswdTestData{
			userInfo: client.User{
				Username: testBasic1Username,
				Password: newPassword,
			},
			expectedResponse: http.StatusOK,
		},
		uPasswdTestData{
			userInfo: client.User{
				Username: testBasic2Username,
				Password: newPassword,
			},
			expectedResponse: http.StatusOK,
		},
	}
	for _, d := range testData {
		requestBody, err := json.Marshal(d.userInfo)
		assert.NoError(t, err)
		request := httptest.NewRequest("POST", "http://user/UpdateUserPassword", bytes.NewBuffer(requestBody))
		request.AddCookie(tokenCookie)
		response := httptest.NewRecorder()
		server.UpdateUserPassword(response, request)
		assert.Equal(t, d.expectedResponse, response.Code)
	}
}

func TestBasicUpdatePassword(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testBasic1Username, Password: testBasic1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	// Attempt to change the password of various users
	newPassword := "newPasswordValue"
	testData := []uPasswdTestData{
		uPasswdTestData{
			userInfo: client.User{
				Username: testAdmin1Username,
				Password: newPassword,
			},
			expectedResponse: http.StatusUnauthorized,
		},
		uPasswdTestData{
			userInfo: client.User{
				Username: testStaff1Username,
				Password: newPassword,
			},
			expectedResponse: http.StatusUnauthorized,
		},

		uPasswdTestData{
			userInfo: client.User{
				Username: testBasic1Username,
				Password: newPassword,
			},
			expectedResponse: http.StatusOK,
		},
		uPasswdTestData{
			userInfo: client.User{
				Username: testBasic2Username,
				Password: newPassword,
			},
			expectedResponse: http.StatusUnauthorized,
		},
	}
	for _, d := range testData {
		requestBody, err := json.Marshal(d.userInfo)
		assert.NoError(t, err)
		request := httptest.NewRequest("POST", "http://user/UpdateUserPassword", bytes.NewBuffer(requestBody))
		request.AddCookie(tokenCookie)
		response := httptest.NewRecorder()
		server.UpdateUserPassword(response, request)
		assert.Equal(t, d.expectedResponse, response.Code)
	}
}
