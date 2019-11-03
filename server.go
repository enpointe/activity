package main

import (
	"log"
	"net/http"

	"github.com/enpointe/activity/views"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	opt := views.DBOptions(clientOptions)
	activityServer, err := views.NewServerService(false, opt, views.DBName("activities"))
	if err != nil {
		log.Panic(err)
	}
	http.HandleFunc("/login", activityServer.Login)
	http.HandleFunc("/logout", activityServer.Logout)
	http.HandleFunc("/user/create", activityServer.CreateUser)
	http.HandleFunc("/user/", activityServer.GetUser)
	http.HandleFunc("users/", activityServer.GetUsers)
	http.ListenAndServe(":8080", nil)
}
