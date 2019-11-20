package controllers

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
	"strings"
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

// CreateUser create a user and add it to our list of known users.
// The POST request should contain a JSON payload that specifies the JSON request
// fields in client.User. The client.User.id field is ignored and is returned
// in the result. The id returned represents the identifier for retrieving
// information about that specific user.
//
// The privileges of the user invoking this method determine whether this operation
// can be performed. A admin privileged user can create a user with any privilege level.
// A staff privileged user can create a staff or a basic privilege level user.
// A basic privilege user can not create any users. If the user doesn't have the
// proper privileges then this operation will fail with http.StatusUnauthorized response.
//
// The JWT cookie, token will be validated to ensure the user is logged into the system
func (s *ServerService) CreateUser(response http.ResponseWriter, request *http.Request) {
	log.Trace("CreateUser request")
	if request.Method != "POST" {
		errorWithJSON(response, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
	claims, httpStatus := validateClaim(response, request)
	if httpStatus != http.StatusOK {
		errorWithJSON(response, http.StatusText(httpStatus), httpStatus)
		return
	}

	// Only allow operation if the user is an staff or administrator
	if !claims.Privilege.Grants(perm.Staff) {
		errorWithJSON(response,
			http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
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
		errorWithJSON(response,
			http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	log.Tracef("Request by %s:%s to create new user %s with perm %s", claims.ID, claims.Username,
		user.Username, user.Privilege)
	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
	defer cancel()
	userService, err := db.NewUserService(s.Database)
	if err != nil {
		errorWithJSON(response, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := userService.Create(ctx, &user)
	if err != nil {
		log.Error("Failed to create user ", err)
		errorWithJSON(response, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Infof("%s:%s created user %s:%s",
		claims.ID, claims.Username, user.Username, id)
	response.Header().Set("content-type", "application/json")
	response.WriteHeader(http.StatusOK)
	// Return the ID a result of creation
	result := struct {
		ID string `json:"id"`
	}{id}
	json.NewEncoder(response).Encode(result)
	return
}

// DeleteUser delete the user specified as an ID as the last element in the URL path.
//
// The privileges of the user invoking this method determine whether this operation
// can be performed.
//
// A admin privileged user can delete a user with any privilege level.
// A staff privileged user can delete a basic privilege level user.
// A basic privilege user can not delete any users.
//
// The JWT cookie, token will be validated to ensure the user is logged into the system
func (s *ServerService) DeleteUser(response http.ResponseWriter, request *http.Request) {
	log.Trace("DeleteUser request")
	if request.Method != "POST" {
		errorWithJSON(response, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
	claims, httpStatus := validateClaim(response, request)
	if httpStatus != http.StatusOK {
		errorWithJSON(response, http.StatusText(httpStatus), httpStatus)
		return
	}

	// Only allow operation if the user is an staff or administrator
	if !claims.Privilege.Grants(perm.Staff) {
		errorWithJSON(response,
			http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Retrieve the id from the URL of the user to delete
	p := request.URL.EscapedPath()
	id := path.Base(p)
	if len(id) == 0 || strings.HasSuffix(p, "/") {
		// Request does not contain requested user
		errorWithJSON(response, "Unable to delete user, no id specified", http.StatusBadRequest)
		return
	}

	// A user can not delete themselves
	if claims.ID == id {
		errorWithJSON(response, "User can not delete themselves", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
	defer cancel()
	userService, err := db.NewUserService(s.Database)
	if err != nil {
		errorWithJSON(response, err.Error(), http.StatusInternalServerError)
		return
	}

	if claims.Privilege == perm.Staff {
		// Check the privileges of the user that the staff privileged user
		// wished to delete.
		userInfo, err := userService.GetByID(ctx, id)
		if err != nil {
			errorWithJSON(response, err.Error(), http.StatusInternalServerError)
			return
		}
		if perm.Convert(userInfo.Privilege) != perm.Basic {
			// Staff user can not delete a Staff or Admin user
			errorWithJSON(response,
				http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	}

	log.Tracef("%s:%s requested delete of user %s", claims.ID, claims.Username, id)
	cnt, err := userService.DeleteUserData(ctx, id)
	if err != nil || cnt == 0 {
		log.Errorf("%s:%s failed to delete user %s, %s",
			claims.ID, claims.Username, id, err)
		if err != nil {
			errorWithJSON(response, err.Error(), http.StatusInternalServerError)
			return
		}
		// user delete failed as no users were deleted
		errorWithJSON(response, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	log.Infof("%s:%s successfully deleted user %s", claims.ID, claims.Username, id)

	// Return a count of the # of entries deleted
	result := struct {
		Count int `json:"deleteCount"`
	}{cnt}
	response.Header().Set("content-type", "application/json")
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
	if request.Method != "GET" {
		errorWithJSON(response, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}

	claims, httpStatus := validateClaim(response, request)
	if httpStatus != http.StatusOK {
		errorWithJSON(response, http.StatusText(httpStatus), httpStatus)
		return
	}

	// Retrieve the user ID from the URL
	p := request.URL.EscapedPath()
	userID := path.Base(p)
	if len(userID) == 0 || strings.HasSuffix(p, "/") {
		// Request does not contain requested user
		errorWithJSON(response,
			"Unable to fetch user data, no user id specified", http.StatusBadRequest)
		return
	}

	// Only allow operation if perm.ADMIN or perm.ADMIN
	// or perm.BASIC and user is requesting details about themselves
	if !claims.Privilege.Grants(perm.Staff) {
		// This operation is allows with perm.Basic if the
		// user is requesting data about themselves
		if claims.ID != userID {
			log.Tracef("User not authorized claims.ID %s != %s", claims.ID, userID)
			errorWithJSON(response,
				http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
	defer cancel()
	userService, err := db.NewUserService(s.Database)
	if err != nil {
		errorWithJSON(response, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := userService.GetByID(ctx, userID)
	if err != nil {
		errorWithJSON(response, err.Error(), http.StatusBadRequest)
		return
	}
	response.Header().Set("content-type", "application/json")
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(user)
}

// GetUsers A GET request that returns information about all known users
func (s *ServerService) GetUsers(response http.ResponseWriter, request *http.Request) {
	log.Trace("GetUsers request")
	if request.Method != "GET" {
		errorWithJSON(response, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}

	claims, httpStatus := validateClaim(response, request)
	if httpStatus != http.StatusOK {
		errorWithJSON(response, http.StatusText(httpStatus), httpStatus)
		return
	}

	// Only allow operation if the user is an administer/staff
	if !claims.Privilege.Grants(perm.Staff) {
		errorWithJSON(response,
			http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 120*time.Second)
	defer cancel()
	userService, err := db.NewUserService(s.Database)
	if err != nil {
		errorWithJSON(response, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := userService.GetAll(ctx)
	if err != nil {
		errorWithJSON(response, err.Error(), http.StatusInternalServerError)
		return
	}
	response.Header().Set("content-type", "application/json")
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(user)
}

// UpdateUserPassword the password for a user
// The POST request should contain a JSON payload that specified the JSON
// request fields used to update the password.
//
// The privileges of the user determine what password update operations can be performed.
// A user can always has the necessary privileges to update their own password.
// A admin privileged user can update the password of any user.
// A staff privileged user can update the password for any basic privilege user.
// A basic privilege user can only update there own password.
//
// The JWT cookie, token will be validated to ensure the user is logged into the system
func (s *ServerService) UpdateUserPassword(response http.ResponseWriter, request *http.Request) {
	log.Trace("UpdateUserPassword request")
	if request.Method != "POST" {
		errorWithJSON(response, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
	claims, httpStatus := validateClaim(response, request)
	if httpStatus != http.StatusOK {
		errorWithJSON(response, http.StatusText(httpStatus), httpStatus)
		return
	}

	var pUpdate client.PasswordUpdate
	err := json.NewDecoder(request.Body).Decode(&pUpdate)
	if err != nil {
		errorWithJSON(response, err.Error(), http.StatusBadRequest)
		return
	}

	userService, err := db.NewUserService(s.Database)
	if err != nil {
		errorWithJSON(response,
			http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
	defer cancel()

	// If the user is attempting to change there own password they must include
	// the current password, otherwise we ignore any value in it
	if pUpdate.ID == claims.ID {
		if len(pUpdate.CurrentPassword) == 0 {
			errorWithJSON(response, "current password must be specified", http.StatusBadRequest)
			return
		}
		// If the user is trying to change there own password, revalidate them
		creds := client.Credentials{
			Username: claims.Username,
			Password: pUpdate.CurrentPassword,
		}
		_, err := userService.Validate(ctx, &creds)
		if err != nil {
			log.Warning("Credentials didn't validate")
			errorWithJSON(response,
				http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	} else {
		// We ignore current password, if provided, clear it
		pUpdate.CurrentPassword = ""

		// Only allow operation if the user is an staff or administrator
		if !claims.Privilege.Grants(perm.Staff) {
			errorWithJSON(response,
				http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Enforce rules on who can perform a password update
		if claims.Privilege == perm.Staff {
			// Staff have restricts on what operations can be performed
			// A staff privileged user can update the password for any basic privilege user.
			// Retrieve information about the user being operated on.
			userInfo, err := userService.GetByID(ctx, pUpdate.ID)
			if err != nil {
				errorWithJSON(response, err.Error(), http.StatusInternalServerError)
				return
			}
			if perm.Convert(userInfo.Privilege) != perm.Basic {
				// Staff user can not modify password credentials of Staff or Admin user
				errorWithJSON(response,
					http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}
	}

	// Update the password for the user
	cnt, err := userService.UpdatePassword(ctx, &pUpdate)
	if err != nil {
		errorWithJSON(response, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("%s:%s updated user password",
		claims.ID, claims.Username)
	// Return a count of the # of entries updated
	result := struct {
		Count int `json:"updateCount"`
	}{cnt}
	response.Header().Set("content-type", "application/json")
	json.NewEncoder(response).Encode(result)
	response.WriteHeader(http.StatusOK)
	return
}
