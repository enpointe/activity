package views

//
// TODO - Create customize error type so that Internal Errors can be distinguished between errors types that
// we want to report StatusOK back for
//

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/enpointe/activity/perm"
)

func errorWithJSON(response http.ResponseWriter, message string, code int) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(code)
	fmt.Fprintf(response, `{ "message": "%q" }`, message)
}

func responseWithJSON(response http.ResponseWriter, json []byte, code int) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(code)
	response.Write(json)
}

// CreateUser create a user and add it to our list of known users
func (s *ServerService) CreateUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	if request.Method != "POST" {
		errorWithJSON(response, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
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
		errorWithJSON(response, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
	defer cancel()
	userService, err := db.NewUserService(s.Database, s.log)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	err = userService.Create(ctx, &user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	response.WriteHeader(http.StatusOK)
}

// GetUser return stored information for a specific user
func (s *ServerService) GetUser(response http.ResponseWriter, request *http.Request) {
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

	// Only allow operation if perm.ADMIN or perm.ADMIN
	// or perm.BASIC and user is requesting details about themselves
	if !claims.Privilege.Grants(perm.Staff) {
		// This operation is allows with perm.Basic if the
		// user is requesting data about themselves
		if claims.Username != username {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
	defer cancel()
	userService, err := db.NewUserService(s.Database, s.log)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	user, err := userService.GetByUsername(ctx, username)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(user)
}

// GetUsers return information about all known users
func (s *ServerService) GetUsers(response http.ResponseWriter, request *http.Request) {
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

	ctx, cancel := context.WithTimeout(context.TODO(), 120*time.Second)
	defer cancel()
	userService, err := db.NewUserService(s.Database, s.log)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	user, err := userService.GetAll(ctx)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(user)
}
