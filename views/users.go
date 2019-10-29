package views

//
// TODO - Create customize error type so that Internal Errors can be distinguished between errors types that
// we want to report StatusOK back for
//

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/enpointe/activity/perm"
)

// CreateUser create a user and add it to our list of known users
func CreateUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	if request.Method != "POST" {
		response.WriteHeader(http.StatusMethodNotAllowed)
		response.Write([]byte(`{ "message": "` + http.StatusText(http.StatusMethodNotAllowed) + `" }`))
		return
	}
	claims, httpStatus := validateClaim(response, request)
	if httpStatus != http.StatusOK {
		response.WriteHeader(httpStatus)
		return
	}

	// Only allow operation if the user is an staff or administrator
	if !claims.Privilege.Grants(perm.Staff) {
		response.WriteHeader(http.StatusUnauthorized)
		return
	}

	var user client.User
	err := json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	config := db.Config{}
	userService, err := db.NewUserService(&config)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	err = userService.CreateUser(&user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	response.WriteHeader(http.StatusOK)
}

// GetUser return stored information for a specific user
func GetUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	if request.Method != "GET" {
		response.WriteHeader(http.StatusMethodNotAllowed)
		response.Write([]byte(`{ "message": "` +
			http.StatusText(http.StatusMethodNotAllowed) + `" }`))
		return
	}

	claims, httpStatus := validateClaim(response, request)
	if httpStatus != http.StatusOK {
		response.WriteHeader(httpStatus)
		return
	}

	// Retreive the username from the URL. usernames are unique
	username := strings.TrimPrefix(request.URL.EscapedPath(), "/activity/user/")
	if len(username) == 0 {
		// Request does not contain requested user
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(`{ "message": "Unable to fetch user data, no user id specified" }`))
		return
	}

	// Only allow operation if the user is an administer/staff
	// or request is being made to retrieves owners record
	if !(claims.Privilege.Grants(perm.Staff) || claims.Username == username) {
		response.WriteHeader(http.StatusUnauthorized)
		return
	}

	config := db.Config{}
	userService, err := db.NewUserService(&config)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	user, err := userService.GetUserByUsername(username)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(user)
}

// GetUsers return information about all known users
func GetUsers(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	if request.Method != "GET" {
		response.WriteHeader(http.StatusMethodNotAllowed)
		response.Write([]byte(`{ "message": "` +
			http.StatusText(http.StatusMethodNotAllowed) + `" }`))
		return
	}

	claims, httpStatus := validateClaim(response, request)
	if httpStatus != http.StatusOK {
		response.WriteHeader(httpStatus)
		return
	}

	// Only allow operation if the user is an administer/staff
	if !claims.Privilege.Grants(perm.Staff) {
		response.WriteHeader(http.StatusUnauthorized)
		return
	}

	config := db.Config{}
	userService, err := db.NewUserService(&config)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	user, err := userService.GetAllUsers()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(user)
}
