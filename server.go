package main

import (
	"net/http"

	"github.com/enpointe/activity/views"
)

func main() {
	activityServer := views.NewServer(views.ListenAddress(":8080"))
	http.HandleFunc("/login", activityServer.Login)
	http.HandleFunc("/logout", activityServer.Logout)
	http.HandleFunc("/user/create", activityServer.CreateUser)
	http.HandleFunc("/user/", activityServer.GetUser)
	http.HandleFunc("users/", activityServer.GetUsers)
	http.ListenAndServeTLS(activityServer, nil)
}
