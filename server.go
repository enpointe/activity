package main

import (
	"net/http"

	"github.com/enpointe/activity/views"
)

func main() {
	http.HandleFunc("/login", views.Login)
	http.HandleFunc("/logout", views.Logout)
	http.HandleFunc("/user/create", views.CreateUser)
	http.HandleFunc("/user/", views.GetUser)
	http.HandleFunc("users/", views.GetUsers)
	http.ListenAndServe(":8080", nil)
}
