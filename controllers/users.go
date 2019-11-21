package controllers

//
// TODO - Create customize error type so that Internal Errors can be distinguished between errors types that
// we want to report StatusOK back for
//

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/enpointe/activity/perm"
	log "github.com/sirupsen/logrus"
)

// APIError Returns a error to the client
type APIError struct {
	ErrorCode    int    `json:"code" example:"400"`
	ErrorMessage string `json:"message" example:"status bad request"`
}

func errorWithJSON(response http.ResponseWriter, message string, code int) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(code)
	error := APIError{
		ErrorCode:    code,
		ErrorMessage: message,
	}
	json.NewEncoder(response).Encode(error)
}

// Identity Used to return the ID of a create operation
type Identity struct {
	ID string `json:"id,unique" example:"5db8e02b0e7aa732afd7fbc4"`
}

// CreateUser create a user and add it to our list of known users.
// The POST request should contain a JSON payload that specifies the JSON request
// fields in client.UserCreate. The id returned represents the identifier for retrieving
// information about that specific user.
//
// The privileges of the user invoking this method determine whether this operation
// can be performed. A admin privileged user can create a user with any privilege level.
// A staff privileged user can create a staff or a basic privilege level user.
// A basic privilege user can not create any users.
//
// The JWT cookie, token will be validated to ensure the user is logged into the system
//
// @Summary Create a user for the activity server
// @Description Create a user for the activity server.
// @Description The privileges of the user invoking this method determine whether this operation
// @Description can be performed. A admin privileged user can create a user with any privilege level.
// @Description A staff privileged user can create a staff or a basic privilege level user.
// @Description A basic privilege user can not create any users.
// @Tags client.UserCreate Identity
// @Security ApiKeyAuth
// @in header
// @name Authorization
// @Param UserCreate body client.UserCreate true "Configuration Data of the user being create"
// @param Authorization header string true "The JWT authorization token acquired at login""
// @Accept  json
// @Produce  json
// @Success 200 {object} Identity
// @Failure 400 {object} APIError "Bad Request"
// @Failure 401 {object} APIError "Unauthorized, if the user lacks privileges to perform operation"
// @Failure 405 {object} APIError "Method Not Allowed"
// @Failure 500 {object} APIError "Internal Server Error"
// @Router /user/create [post]
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

	var user client.UserCreate
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
	result := Identity{id}
	json.NewEncoder(response).Encode(result)
	return
}

// DeleteCount the
type DeleteCount struct {
	Count int `json:"deleteCount" example:"1"`
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
// The JWT cookie, token will be validated to ensure the user is logged into the systemgodoc
// @Summary Delete a user from the activity server
// @Description Delete a user for the given ID.
// @Description The privileges of the user invoking this method determine whether this operation
// @Description can be performed.
// @Description A admin privileged user can delete a user with any privilege level.
// @Description A staff privileged user can delete a basic privilege level user.
// @Description A basic privilege user can not delete any users.
// @Tags DeleteCount
// @Security ApiKeyAuth
// @in header
// @name Authorization
// @Param user_id path string true "ID of the user to delete"
// @param Authorization header string true "The JWT authorization token acquired at login"
// @Accept  json
// @Produce  json
// @Success 200 {object} DeleteCount "Number of items deleted"
// @Failure 400 {object} APIError "Bad Request"
// @Failure 401 {object} APIError "Unauthorized, if the user lacks privileges to perform operation"
// @Failure 404 {object} APIError "Not Found, if the ID of the user to delete is not found"
// @Failure 405 {object} APIError "Method Not Allowed"
// @Failure 500 {object} APIError "Internal Server Error"
// @Router /user/delete/{user_id} [delete]
func (s *ServerService) DeleteUser(response http.ResponseWriter, request *http.Request) {
	log.Trace("DeleteUser request")
	if request.Method != "DELETE" {
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
	result := DeleteCount{cnt}
	response.Header().Set("content-type", "application/json")
	json.NewEncoder(response).Encode(result)
	response.WriteHeader(http.StatusOK)
}

// GetUser return stored information for a specific user ID contained as the last path
// in the URL GET request. If the URL of the request is "user/5dc2ee5a567855de21f1070a" then
// "5dc2ee5a567855de21f1070a" value will be the ID used to retrieve information for.
//
// The privileges of the user invoking this method are used to determine what requests
// can be satisfied.
// A admin and staff privileged user can fetch details about any user.
// A basic privilege user can only fetch details about themselves.
//
// @Summary Get information about the specified user
// @Description Get the client.UserInfo data for the specified user ID.
// @Description The privileges (client.UserInfo.Privilege) of the user invoking this
// @Description method are used to determine what requests can be satisfied.
// @Description A admin and staff privileged user can fetch details about any user.
// @Description A basic privilege user can only fetch details about themselves.
// @Tags client.UserInfo
// @Security ApiKeyAuth
// @in header
// @name Authorization
// @Param user_id path string true "ID of the User to fetch details about"
// @param Authorization header string true "The JWT authorization token acquired at login"
// @Accept  json
// @Produce  json
// @Success 200 {object} client.UserInfo
// @Failure 400 {object} APIError "Bad Request"
// @Failure 401 {object} APIError "Unauthorized"
// @Failure 404 {object} APIError "Not Found"
// @Failure 405 {object} APIError "Method Not Allowed"
// @Failure 500 {object} APIError "Internal Server Error"
// @Router /user/{user_id} [get]
func (s *ServerService) GetUser(response http.ResponseWriter, request *http.Request) {
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
		errorWithJSON(response, err.Error(), http.StatusNotFound)
		return
	}
	response.Header().Set("content-type", "application/json")
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(user)
}

// GetUsers A GET request that returns information about all known users.
// Only admin and staff privileged users can perform this operation.
//
// The JWT cookie, token will be validated to ensure the user is logged into the system.
//
// @Summary Get user information for all users
// @Description Get client.UserInfo data for all known users.
// @Description Only admin and staff privileged users can perform this operation.
// @Tags client.UserInfo
// @Security ApiKeyAuth
// @in header
// @name Authorization
// @param Authorization header string true "The JWT authorization token acquired at login"
// @Accept  json
// @Produce  json
// @Success 200 {array} client.UserInfo
// @Failure 400 {object} APIError "Bad Request"
// @Failure 401 {object} APIError "Unauthorized"
// @Failure 404 {object} APIError "Not Found"
// @Failure 405 {object} APIError "Method Not Allowed"
// @Failure 500 {object} APIError "Internal Server Error"
// @Router /users/ [get]
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

// UpdateResults the results of the update operation
type UpdateResults struct {
	Count int `json:"updateCount" example:"1"`
}

// UpdateUserPassword the password for a user
// The PATCH request should contain a JSON payload that specified the JSON
// request fields used to update the password.
//
// The privileges of the user determine what password update operations can be performed.
// A user always has the necessary privileges to update their own password.
// A admin privileged user can update the password of any user.
// A staff privileged user can update the password for any basic privilege user.
// A basic privilege user can only update there own password.
//
// The JWT cookie, token will be validated to ensure the user is logged into the system
//
// @Summary Update the password for a user
// @Description Updates the password for a given user. If a user is changing there own
// @Description password the current password must be specified.
// @Description
// @Description The privileges of the user determine what password update operations can be performed.
// @Description A user always has the necessary privileges to update their own password.
// @Description A admin privileged user can update the password of any user.
// @Description A staff privileged user can update the password for any basic privilege user.
// @Description A basic privilege user can only update there own password.
//
// @Tags client.UserPassword
// @Security ApiKeyAuth
// @in header
// @name Authorization
// @Param PasswordUpdate body client.PasswordUpdate true "Parameters for updating the specified users password"
// @param Authorization header string true "The JWT authorization token acquired at login"
// @Accept  json
// @Produce  json
// @Success 200 {object} UpdateResults
// @Failure 400 {object} APIError "Bad Request"
// @Failure 401 {object} APIError "Unauthorized"
// @Failure 404 {object} APIError "Not Found"
// @Failure 405 {object} APIError "Method Not Allowed"
// @Failure 500 {object} APIError "Internal Server Error"
// @Router /user/updatePasswd [patch]
func (s *ServerService) UpdateUserPassword(response http.ResponseWriter, request *http.Request) {
	log.Trace("UpdateUserPassword request")
	if request.Method != "PATCH" {
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
	result := UpdateResults{cnt}
	response.Header().Set("content-type", "application/json")
	json.NewEncoder(response).Encode(result)
	response.WriteHeader(http.StatusOK)
	return
}
