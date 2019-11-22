
# Activity 
<p align="center">
  <a href="https://goreportcard.com/report/github.com/enpointe/activity"><img src="https://goreportcard.com/badge/github.com/enpointe/activity"></a>

**Version 2.0**

- [Status](#status)
- [Work Outstanding](#work-outstanding)
- [Requirements](#requirements)
- [Building](#building)
- [Running](#running)
- [REST API Interface](#rest-api-interface)
- [Project Structure](#project-structure)
- [References](#references)

This is a rework/expansion of the version 1.0 implementation of this activity logger which was written 
as a take home project for a job interview.

This project is primarily intended as a exercise in Go Programing

In order to make the origional project a bit more interesting the overall project has been expanded to include

* Add methods to support CRUD operations
* Add backend storage using MongoDB
    * Three database tables are being added
        * Users  - table to hold user information (username, password, privileges)
        * Exercise - table to hold a list of exercises and there descriptions
        * Log - table to log time spent on various exercises by a specific user
* Add a front end http interface to expose CRUD JSON methods
    * Add user authentication
    * Add some form of user priviledges

-------------------------
# Status

The current state of this project

## Completed
Current status of work completed so far:

* Basic Templates for database tables has been created
* Initial model, controller layout has been created.
    * model db/client layout is being used to maintain a clear seperation of data objects
* Initial http interfaces for user have been created
* Basic login/logout with JWT authentication has been implemented.
    * JWT token stored as a cookie 

## Work outstanding

* Add ability to log exercise workouts
* Improve unit test
    * Investigate how to mock database calls
* Create custom error types so that more realistic https status code can be returned when a error occurs at the database level due to bad data in request
* Add database configuration for security
* Add mechanism for prepopulating database with Exercises
    * mongoimport is available via [mongo tools](https://github.com/mongodb/mongo-tools)
* Examine whether view code should have some context cancel in it. See [Stack Overflow question](https://stackoverflow.com/questions/47179024/how-to-check-if-a-request-was-cancelled)
* JWT secret key needs to be configurable.
    * Consider creating configuration file for this
* Need to create an initial admin user in order for http interfaces to function.
    * Consider creating a cli interface for this
* perm authorization is a bit comberson. Consider adding a simple RBAC authorization on methods or 
    * [Casbin](https://github.com/casbin/casbin)
    * [goRBAC](https://github.com/mikespook/gorbac)
* Auditing - Understand the best method for recording audit level changes


# Issues 

[![Open issues](https://img.shields.io/github/issues/enpointe/activity)](https://github.com/enpointe/activity) [![Closed issues](https://img.shields.io/github/issues-closed/enpointe/activity)](https://github.com/enpointe/activity/issues?q=is%3Aissue+is%3Aclosed) [![Open PRs](https://img.shields.io/github/issues-pr/enpointe/activity)](https://github.com/enpointe/activity/pulls) [![Closed PRs](https://img.shields.io/github/issues-pr-closed/enpointe/activity)](https://github.com/enpointe/activity/pulls?q=is%3Apr+is%3Aclosed)

Most issues of this project are currently being tracked via TODO in the code. Larger items that need to be done are currently tracked in Work Outstanding. Gradually all these issues will be moved to Issues Tracking as the project formalizes a bit more.

# Requirements

As this tools uses MongoDB it will be necessary to install and start the mongod server before executing the application or
running any tests. Currently the server needs to be accessible via 'mongodb://localhost:27017'. This change as
features and options are added to this project.

Start the MongoDB database
```
$ mongod
```

# Building

A convience Makefile is provided to build the application.

```
$ make
```

Regenerating or changing the swagger documentation requires the installation of swaggo for information
on the installation of this tool see 
[Swaggo Swag Getting Started](https://github.com/swaggo/swag#getting-started). The 'swag' command
must be findable in your PATH.

# Running

## Starting the Activity Server

1. Ensure the MongoDB server is running and available on the URL, mongodb://localhost:27017
2. If starting the server for the first time it will be necessary to create a Administrative user, "admin".  This
can be done via the "--admin <password>" flag to the activity server. Where password represent the password to
set for the administrative user.
3. Start the activity server


```
$ activity 
```

or 

```
$ activity -admin <password>
```

## REST API Interface

The REST API HTTP interface for this module is documented using swagger. Once the activity server is started the 
REST API documentation can be viewed via

[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

**Note** The current implementation stored the authorization token in a cookie.  The
swagger documentation currently identifies it as being stored in the header. The underling
implementation will be changed to match the documentation. For now the curl requests 
outlined in the documented will have to be changed to use -b "authorization" flag.

The following HTTP REST API methods are currently available:
```
http://localhost:8080/login
http://localhost:8080/logout
http://localhost:8080/user/{user_id}
http://localhost:8080/users
http://localhost:8080/user/create
http://localhost:8080/user/delete/{user_id}
http://localhost:8080/user/updatePasswd
```

# Project Structure

This project is laid out as a Go module in a hierachy to support the notion of Model-View-Container architecture. 

```
├── models                      // Models for our application
│   ├── client                  // Model for client
│   │   ├── credentials.go      // Login Credentials API
│   │   ├── exercise.go         // Exercise API
│   │   ├── user.go             // User API
│   ├── db                      // APIs for access the database
│   │   ├── exercise.go         // Model for exercise collection
│   │   ├── exercise_service.go // APIs for exercise collection
│   │   ├── user.go             // Model for users collection
│   │   ├── user_service.go     // APIs for user collection
├── perm                        // Permission model for method access control
│   └── priv.go                 // Permissions level used for access control
├── controllers                 // Controller APIs
│       └── claims.go           // JWT claims
│       └── login.go            // HTTP login REST API interface
│       └── logout.go           // HTTP logout REST API interface
│       └── server_service.go   // HTTP Server Service
│       └── users.go            // HTTP REST API interface for interacting with the user model
├── scripts                     // Scripts
│   └── start-dev-container.sh  // Docker script for starting up development environment
└── server.go                   // Server application

```

# References

References used during the development of this project

* [Make yourself a Go web server with MongoDb](https://medium.com/hackernoon/make-yourself-a-go-web-server-with-mongodb-go-on-go-on-go-on-48f394f24e)
* [Implementing JWT based authentication in Golang](https://www.sohamkamani.com/blog/golang/2019-01-01-jwt-authentication/)
* [Build and Deploy a secure REST API with Go, Postgresql, JWT and GORM](https://medium.com/@adigunhammedolalekan/build-and-deploy-a-secure-rest-api-with-go-postgresql-jwt-and-gorm-6fadf3da505b)
* [Setting Up Swagger Docs for Golang API](https://towardsdatascience.com/setting-up-swagger-docs-for-golang-api-8d0442263641)