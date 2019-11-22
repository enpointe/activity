package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/enpointe/activity/perm"
	"github.com/stretchr/testify/assert"
	"syreclabs.com/go/faker"
)

type uCreateTestData struct {
	userInfo         client.UserCreate
	expectedResponse int
}

func testCreateUser(t *testing.T, creds client.Credentials, testData []uCreateTestData) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	for _, d := range testData {
		t.Run(fmt.Sprintf("User-%s:%s:%s", d.userInfo.Username, d.userInfo.Password, d.userInfo.Privilege),
			func(t *testing.T) {
				requestBody, err := json.Marshal(d.userInfo)
				assert.NoError(t, err)
				request := httptest.NewRequest(http.MethodPost, "http://users", bytes.NewBuffer(requestBody))
				request.AddCookie(tokenCookie)
				response := httptest.NewRecorder()
				server.CreateUser(response, request, nil)
				assert.Equalf(t, d.expectedResponse, response.Code,
					"%s attempted to create user %s password: %s, privilege: %s, expected '%s' got '%s'",
					creds.Username, d.userInfo.Username, d.userInfo.Password, d.userInfo.Privilege,
					http.StatusText(d.expectedResponse), http.StatusText(response.Code))
			})
	}
}

// TestCreateValidationFailures test for validation errors, invalid data in UserCreate structure
func TestCreateValidationFailures(t *testing.T) {
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	testData := []uCreateTestData{
		uCreateTestData{
			userInfo: client.UserCreate{ // Invalid username
				Username:  "a",
				Password:  faker.Internet().Password(db.PasswordMinLength, 20),
				Privilege: perm.Basic.String(),
			},
			expectedResponse: http.StatusUnprocessableEntity,
		},
		uCreateTestData{
			userInfo: client.UserCreate{ // Invalid password
				Username:  faker.Internet().UserName(),
				Password:  "",
				Privilege: perm.Basic.String(),
			},
			expectedResponse: http.StatusUnprocessableEntity,
		},
		uCreateTestData{
			userInfo: client.UserCreate{ // Invalid privilege
				Username:  faker.Internet().UserName(),
				Password:  faker.Internet().Password(db.PasswordMinLength, 30),
				Privilege: "Unknown",
			},
			expectedResponse: http.StatusUnprocessableEntity,
		},
	}
	testCreateUser(t, creds, testData)
}

func TestAdminCreateUser(t *testing.T) {
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	testData := []uCreateTestData{
		uCreateTestData{
			userInfo: client.UserCreate{ // Create admin privilege user
				Username:  faker.Internet().UserName(),
				Password:  faker.Internet().Password(db.PasswordMinLength, 30),
				Privilege: perm.Admin.String(),
			},
			expectedResponse: http.StatusCreated,
		},
		uCreateTestData{
			userInfo: client.UserCreate{ // Create staff privilege user
				Username:  faker.Internet().UserName(),
				Password:  faker.Internet().Password(db.PasswordMinLength, 30),
				Privilege: perm.Staff.String(),
			},
			expectedResponse: http.StatusCreated,
		},
		uCreateTestData{
			userInfo: client.UserCreate{ // Create basic privilege user
				Username:  faker.Internet().UserName(),
				Password:  faker.Internet().Password(db.PasswordMinLength, 30),
				Privilege: perm.Basic.String(),
			},
			expectedResponse: http.StatusCreated,
		},
		uCreateTestData{
			userInfo: client.UserCreate{ // Attempt to create an existing user
				Username:  testAdmin1Username,
				Password:  testAdmin1UserPassword,
				Privilege: perm.Admin.String(),
			},
			expectedResponse: http.StatusConflict,
		},
	}
	testCreateUser(t, creds, testData)
}

func TestStaffCreateUser(t *testing.T) {
	creds := client.Credentials{Username: testStaff1Username, Password: testStaff1UserPassword}
	testData := []uCreateTestData{
		uCreateTestData{
			userInfo: client.UserCreate{ // Attempt to create admin privilege user
				Username:  faker.Internet().UserName(),
				Password:  faker.Internet().Password(db.PasswordMinLength, 30),
				Privilege: perm.Admin.String(),
			},
			expectedResponse: http.StatusForbidden,
		},
		uCreateTestData{
			userInfo: client.UserCreate{ // Create staff privilege user
				Username:  faker.Internet().UserName(),
				Password:  faker.Internet().Password(db.PasswordMinLength, 30),
				Privilege: perm.Staff.String(),
			},
			expectedResponse: http.StatusCreated,
		},
		uCreateTestData{
			userInfo: client.UserCreate{ // Create basic privilege user
				Username:  faker.Internet().UserName(),
				Password:  faker.Internet().Password(db.PasswordMinLength, 30),
				Privilege: perm.Basic.String(),
			},
			expectedResponse: http.StatusCreated,
		},
	}
	testCreateUser(t, creds, testData)
}

func TestBasicCreateUser(t *testing.T) {
	creds := client.Credentials{Username: testBasic1Username, Password: testBasic1UserPassword}
	testData := []uCreateTestData{
		uCreateTestData{
			userInfo: client.UserCreate{ // Create admin privilege user
				Username:  faker.Internet().UserName(),
				Password:  faker.Internet().Password(db.PasswordMinLength, 30),
				Privilege: perm.Admin.String(),
			},
			expectedResponse: http.StatusForbidden,
		},
		uCreateTestData{
			userInfo: client.UserCreate{ // Create staff privilege user
				Username:  faker.Internet().UserName(),
				Password:  faker.Internet().Password(db.PasswordMinLength, 30),
				Privilege: perm.Staff.String(),
			},
			expectedResponse: http.StatusForbidden,
		},
		uCreateTestData{
			userInfo: client.UserCreate{ // Create basic privilege user
				Username:  faker.Internet().UserName(),
				Password:  faker.Internet().Password(db.PasswordMinLength, 30),
				Privilege: perm.Basic.String(),
			},
			expectedResponse: http.StatusForbidden,
		},
	}
	testCreateUser(t, creds, testData)
}
