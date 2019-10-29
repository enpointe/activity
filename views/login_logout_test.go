package views_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/enpointe/activity/views"
)

// TODO: Need to figure out a way to prepopulate the database and control which
// database to use for tests at this level.  Possible solutions are
//
// https://medium.com/better-programming/unit-testing-code-using-the-mongo-go-driver-in-golang-7166d1aa72c0
// https://medium.com/@mvmaasakkers/writing-integration-tests-with-mongodb-support-231580a566cd

func TestLoginLogout(t *testing.T) {
	body := strings.NewReader(`{"username": "admin", "password": "2BadCats@"}`)
	request := httptest.NewRequest("POST", "http://login", body)

	response := httptest.NewRecorder()

	views.Login(response, request)
	assert.Equal(t, response.Code, http.StatusOK)
}
