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

func TestDeleteUser(t *testing.T) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	// Test missing token
	request := httptest.NewRequest(http.MethodDelete, "http://users/"+testBasic1ID, nil)
	response := httptest.NewRecorder()
	ps := httprouter.Params{
		httprouter.Param{
			Key:   "id",
			Value: testBasic1ID,
		},
	}
	server.DeleteUser(response, request, ps)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

type testDeleteData struct {
	id               string
	expectedResponse int
}

// Help function for testing delete using different user and permission combinations
func deleteTest(t *testing.T, creds client.Credentials, deleteData []testDeleteData) {
	server := setup(t, testMultiUserFilenameJSON)
	defer teardown(t, server)
	tokenCookie := login(t, server, creds)
	defer logout(t, server, tokenCookie)

	for _, d := range deleteData {
		t.Run(fmt.Sprintf("ID-%s", d.id),
			func(t *testing.T) {
				request := httptest.NewRequest(http.MethodDelete, "http://user/Delete/"+d.id, nil)
				request.AddCookie(tokenCookie)
				response := httptest.NewRecorder()
				ps := httprouter.Params{
					httprouter.Param{
						Key:   "id",
						Value: d.id},
				}
				server.DeleteUser(response, request, ps)
				assert.Equalf(t, d.expectedResponse, response.Code,
					"%s attempted to delete user ID %s, expected '%s' got '%s'", creds.Username, d.id,
					http.StatusText(d.expectedResponse), http.StatusText(response.Code))
			})
	}
}

// Test deletion when user has admin level privileges
func TestDeleteAdminPrivileges(t *testing.T) {
	creds := client.Credentials{Username: testAdmin1Username, Password: testAdmin1UserPassword}
	testData := []testDeleteData{
		testDeleteData{testAdmin1ID, http.StatusBadRequest}, // Can't delete yourself
		testDeleteData{testStaff1ID, http.StatusOK},
		testDeleteData{testBasic1ID, http.StatusOK},
		testDeleteData{testBasic1ID, http.StatusBadRequest}, // Attempt to delete the same user
		testDeleteData{"", http.StatusBadRequest},           // Attempt to delete without specifying user
	}
	deleteTest(t, creds, testData)
}

// Test deletion when user has staff level privileges
func TestDeleteStaffPrivileges(t *testing.T) {
	creds := client.Credentials{Username: testStaff1Username, Password: testStaff1UserPassword}
	testData := []testDeleteData{
		testDeleteData{testAdmin1ID, http.StatusForbidden},  // Delete of admin not allowed
		testDeleteData{testStaff1ID, http.StatusBadRequest}, // Can't delete yourself
		testDeleteData{testStaff2ID, http.StatusForbidden},
		testDeleteData{testBasic1ID, http.StatusOK},
	}
	deleteTest(t, creds, testData)
}

// Test to ensure basic privilege user can not perform delete action
func TestDeleteBasicPrivileges(t *testing.T) {
	creds := client.Credentials{Username: testBasic1Username, Password: testBasic1UserPassword}
	testData := []testDeleteData{
		testDeleteData{testAdmin1ID, http.StatusForbidden},
		testDeleteData{testStaff1ID, http.StatusForbidden},
		testDeleteData{testBasic1ID, http.StatusForbidden},
	}
	deleteTest(t, creds, testData)
}
