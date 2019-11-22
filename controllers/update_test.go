package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

type uPasswdTestData struct {
	userInfo         client.PasswordUpdate
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
		testData{ // Request is fine but no JSON data
			method:           http.MethodPatch,
			testUser:         testBasic1ID,
			expectedResponse: http.StatusUnsupportedMediaType,
		},
	}
	for _, data := range testInput {
		t.Run(fmt.Sprintf("%s-UpdatePassword-ID-%s", creds.Username, data.testUser),
			func(t *testing.T) {
				url := "http://user/Update/" + data.testUser
				request := httptest.NewRequest(data.method, url, nil)
				request.AddCookie(tokenCookie)
				response := httptest.NewRecorder()
				ps := httprouter.Params{
					httprouter.Param{
						Key:   "id",
						Value: data.testUser,
					},
				}
				server.UpdateUserPassword(response, request, ps)
				assert.Equalf(t, data.expectedResponse, response.Code, "%s %s", data.method, url)
			})
	}

	// Test no token associated with request
	request := httptest.NewRequest("PATCH", "http://user/Update/"+testBasic1ID, nil)
	response := httptest.NewRecorder()
	ps := httprouter.Params{
		httprouter.Param{
			Key:   "id",
			Value: testBasic1ID,
		},
	}
	server.UpdateUserPassword(response, request, ps)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func testUserUpdatePassword(t *testing.T, creds client.Credentials, testData []uPasswdTestData) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	for _, d := range testData {
		t.Run(fmt.Sprintf("%s-UpdatePassword-ID-%s", creds.Username, d.userInfo.ID),
			func(t *testing.T) {
				requestBody, err := json.Marshal(d.userInfo)
				assert.NoError(t, err)
				request := httptest.NewRequest(http.MethodPatch, "http://users/"+d.userInfo.ID, bytes.NewBuffer(requestBody))
				request.AddCookie(tokenCookie)
				response := httptest.NewRecorder()
				ps := httprouter.Params{
					httprouter.Param{
						Key:   "id",
						Value: d.userInfo.ID,
					},
				}
				server.UpdateUserPassword(response, request, ps)
				assert.Equalf(t, d.expectedResponse, response.Code,
					"%s attempted to change password for %s to %s, expected '%s' got '%s'",
					creds.Username, d.userInfo.ID, d.userInfo.NewPassword,
					http.StatusText(d.expectedResponse), http.StatusText(response.Code))
			})
	}
}

func TestAdminUpdatePassword(t *testing.T) {
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}

	// Attempt to change the password of various users
	newPassword := "newPasswordValue"
	testData := []uPasswdTestData{
		uPasswdTestData{ // Fails admin1 current password not specified
			userInfo: client.PasswordUpdate{
				ID:          testAdmin1ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusBadRequest,
		},
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:              testAdmin1ID,
				NewPassword:     newPassword,
				CurrentPassword: testAdmin1UserPassword,
			},
			expectedResponse: http.StatusOK,
		},
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:          testStaff1ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusOK,
		},
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:          testStaff2ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusOK,
		},
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:          testBasic1ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusOK,
		},
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:          testBasic2ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusOK,
		},
	}
	testUserUpdatePassword(t, creds, testData)
}

func TestStaffUpdatePassword(t *testing.T) {
	creds := client.Credentials{Username: testStaff1Username, Password: testStaff1UserPassword}

	// Attempt to change the password of various users
	newPassword := "newPasswordValue"
	testData := []uPasswdTestData{
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:          testAdmin1ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusForbidden,
		},
		uPasswdTestData{ // Fails because current password not specified
			userInfo: client.PasswordUpdate{
				ID:          testStaff1ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusBadRequest,
		},
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:              testStaff1ID,
				NewPassword:     newPassword,
				CurrentPassword: testStaff1UserPassword,
			},
			expectedResponse: http.StatusOK,
		},
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:          testStaff2ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusForbidden,
		},
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:          testBasic1ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusOK,
		},
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:          testBasic2ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusOK,
		},
	}
	testUserUpdatePassword(t, creds, testData)
}

func TestBasicUpdatePassword(t *testing.T) {
	creds := client.Credentials{Username: testBasic1Username, Password: testBasic1UserPassword}

	// Attempt to change the password of various users
	newPassword := "newPasswordValue"
	testData := []uPasswdTestData{
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:          testAdmin1ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusForbidden,
		},
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:          testStaff1ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusForbidden,
		},
		uPasswdTestData{ // Fails since current password not specified
			userInfo: client.PasswordUpdate{
				ID:          testBasic1ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusBadRequest,
		},
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:              testBasic1ID,
				NewPassword:     newPassword,
				CurrentPassword: testBasic1UserPassword,
			},
			expectedResponse: http.StatusOK,
		},
		uPasswdTestData{
			userInfo: client.PasswordUpdate{
				ID:          testBasic2ID,
				NewPassword: newPassword,
			},
			expectedResponse: http.StatusForbidden,
		},
	}
	testUserUpdatePassword(t, creds, testData)
}
