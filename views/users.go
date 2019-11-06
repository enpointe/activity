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
	"path"
	"time"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/enpointe/activity/perm"
	log "github.com/sirupsen/logrus"
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
// The POST request should contain a JSON payload that specifies the JSON request
// fields in client.User. The client.User.id field is ignored and is returned
// in the result. The id returned represents the identifier for retrieving
// information about that specific user.
//
// Restrictions:
// A admin privilege user can create a user with any privilege level
// A staff privilege user can create staff or basic privilege level user
// A basic privilege user can not create any users
func (s *ServerService) CreateUser(response http.ResponseWriter, request *http.Request) {
	log.Trace("CreateUser request")
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

	if claims.Privilege == perm.Staff && user.Privilege == perm.Admin.String() {
		// Staff level user can not create an admin level user
		log.Warnf("%s:%s attempted to create a administrator level user",
			claims.ID, claims.Username)
		response.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Tracef("Request by %s:%s to create new user %s with perm %s", claims.ID, claims.Username,
		user.Username, user.Privilege)
	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
	defer cancel()
	userService, err := db.NewUserService(s.Database)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	id, err := userService.Create(ctx, &user)
	if err != nil {
		log.Error("Failed to create user ", err)
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	log.Infof("%s:%s created user %s:%s",
		claims.ID, claims.Username, user.Username, id)
	response.WriteHeader(http.StatusOK)
	// Return the ID a result of creation
	result := struct {
		ID string `json:"id"`
	}{id}
	json.NewEncoder(response).Encode(result)
	return
}

// DeleteUser delete the user specified in the URL request
func (s *ServerService) DeleteUser(response http.ResponseWriter, request *http.Request) {
	log.Trace("DeleteUser request")
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

	// Retreive the id from the URL of the user to delete
	id := path.Base(request.URL.EscapedPath())
	if len(id) == 0 {
		// Request does not contain requested user
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(`{ "message": "Unable to delete user, no id specified" }`))
		return
	}

	// TODO a staff user should be able to delete a admin user
	// In order to accomplish this we'll need to fetch the user
	// first before deleting.  Need to introduce a transaction to
	// make this operation safe.

	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
	defer cancel()
	userService, err := db.NewUserService(s.Database)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	log.Tracef("%s:%s requested delete of user %s", claims.ID, claims.Username, id)
	err = userService.DeleteUserData(ctx, id)
	if err != nil {
		log.Errorf("%s:%s failed to delete user %s, %s",
			claims.ID, claims.Username, id, err)
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	log.Infof("%s:%s successfully deleted user %s", claims.ID, claims.Username, id)
	// Return a count of the # of entries deleted
	result := struct {
		Count int `json:"deletedEntries"`
	}{1}
	json.NewEncoder(response).Encode(result)
	response.WriteHeader(http.StatusOK)
}

// GetUser return stored information for a specific user ID contained as the last path
// in the URL GET request.
//
// If the URL of the request is "user/5dc2ee5a567855de21f1070a" then
// "5dc2ee5a567855de21f1070a" value will be the ID used to retrieve information for.
func (s *ServerService) GetUser(response http.ResponseWriter, request *http.Request) {
	log.Trace("GetUser request")
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

	// Retreive the user ID from the URL
	userID := path.Base(request.URL.EscapedPath())
	if len(userID) == 0 {
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
		if claims.ID != userID {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
	defer cancel()
	userService, err := db.NewUserService(s.Database)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	user, err := userService.GetByID(ctx, userID)
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
	log.Trace("GetUsers request")
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
	userService, err := db.NewUserService(s.Database)
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
