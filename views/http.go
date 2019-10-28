package views

//
// TODO - Create customize error type so that Internal Errors can be distinguished between errors types that
// we want to report StatusOK back for
//

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/enpointe/activity/models/db"
	"github.com/enpointe/activity/models/server"
)

type requestID struct {
	ID string `json:"ID"`
}

// func IsAdminUser(token string, authService AuthService) int {
// 	userObject := authService.AuthenticateToken(token)
// 	return userObject.IsAdmin() || userObject.IsRoot()
//   }

// Login interface for allowing the user to acquire authorization
// to execute methods for this application
func Login(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	if request.Method != "POST" {
		log.Printf("unauthorized GET login attempt")
		response.WriteHeader(http.StatusUnauthorized)
		return
	}

	var creds server.Credentials
	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil {
		log.Printf("invalid log attempt, bad payload: %s", request.Body)
		response.WriteHeader(http.StatusUnauthorized)
		return
	}
}

// CreateUser create a user and add it to our list of known users
func CreateUser(response http.ResponseWriter, request *http.Request) {
	//ctx := request.Context()
	response.Header().Set("content-type", "application/json")
	if request.Method != "POST" {
		response.WriteHeader(http.StatusMethodNotAllowed)
		response.Write([]byte(`{ "message": "` + http.StatusText(http.StatusMethodNotAllowed) + `" }`))
		return
	}
	var user server.User
	err := json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	userService, err := db.NewUserService()
	err = userService.CreateUser(&user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	response.WriteHeader(http.StatusOK)
}

// // GetUser return stored information for a specific user
// func GetUser(response http.ResponseWriter, request *http.Request) {
// 	response.Header().Set("content-type", "application/json")
// 	if request.Method != "GET" {
// 		response.WriteHeader(http.StatusMethodNotAllowed)
// 		response.Write([]byte(`{ "message": "` +
// 			http.StatusText(http.StatusMethodNotAllowed) + `" }`))
// 		return
// 	}

// 	// Retreive the username from the URL. usernames are unique
// 	username := strings.TrimPrefix(request.URL.EscapedPath(), "/activity/user/")
// 	if len(username) == 0 {
// 		// Request does not contain requested user
// 		response.WriteHeader(http.StatusBadRequest)
// 		response.Write([]byte(`{ "message": "Unable to fetch user data, no user id specified" }`))
// 		return
// 	}

// 	userService, err := db.NewUserService()
// 	user, err := userService.GetUserByUsername(username)
// 	if err != nil {
// 		response.WriteHeader(http.StatusInternalServerError)
// 		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
// 		return
// 	}

// 	response.WriteHeader(http.StatusOK)
// 	json.NewEncoder(response).Encode(user)
// }
