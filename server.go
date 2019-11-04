package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/enpointe/activity/views"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	dbURI := flag.String("dbURI", "mongodb://localhost:27017",
		"URI used to connect to the mongo database, default is mongodb://localhost:27017")
	adminPasswd := flag.String("admin", "", "Create an admin user and assign it the specified password")
	flag.Parse()

	clientOptions := options.Client().ApplyURI(*dbURI)
	sOptions := []views.ServerOption{views.DBOptions(clientOptions)}
	if len(*adminPasswd) > 0 {
		sOptions = append(sOptions, views.CreateAdminUser([]byte(*adminPasswd)))
	}

	sOptions = append(sOptions, views.CreateAdminUser([]byte("changeMe")))
	activityServer, err := views.NewServerService(false, sOptions...)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	http.HandleFunc("/login", activityServer.Login)
	http.HandleFunc("/logout", activityServer.Logout)
	http.HandleFunc("/user/create", activityServer.CreateUser)
	http.HandleFunc("/user/", activityServer.GetUser)
	http.HandleFunc("users/", activityServer.GetUsers)
	http.ListenAndServe(":8080", nil)
}
