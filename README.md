
# Version 2.0 (In development)

This is a rework/expansion of the version 1.0 implementation of this activity logger which was written 
as a take home project for a job interview.

This project is primarily intended as a exercise in Go Programing

In order to make the origional project a bit more interesting the overall project has been expanded to include

* Add backend storage using MongoDB
    * Three database tables are being added
        * Users  - table to hold user information (username, password, privileges)
        * Exercise - table to hold a list of exercises and there descriptions
        * Log - table to log time spent on various exercises by a specific user
* Add a front end http interface to access the JSON methods
    * Add user authentication
    * Add some form of user priviledges

# Status

Current status of work completed so far:

* Basic Templates for database tables has been created
* Initial model, view, controller layout has been created.
    * model db/client layout is being used to maintain a clear seperation of data objects
* Initial http interfaces for user have been created
* Basic login/logout with JWT authentication has been implemented.
    * JWT token stored as a cookie in the session

# Work outstanding

* Improve unit test
    * Investigate how to mock database calls
* Integration Level testing
* Create custom error types so that more realistic https status code can be returned when a error occurs at the database level due to bad data in request
* Add database configuration for security
* Add mechanism for prepopulating database with Exercises
* Examine whether view code should have some context cancel in it. See [Stack Overflow question](https://stackoverflow.com/questions/47179024/how-to-check-if-a-request-was-cancelled)
* JWT secret key needs to be configurable.
    * Consider creating a cli interface for this
* Need to create an initial admin user in order for http interfaces to function.
    * Consider creating a cli interface for this
* perm authorization is a bit comberson. Consider adding a simple RBAC authorization on methods or 
    * [Casbin](https://github.com/casbin/casbin)
    * [goRBAC](https://github.com/mikespook/gorbac)
* models/db - user_service.go - DeleteUserData. This routine needs to remove all
   data associated with the user. This will need to be done as a transaction.
   [See Mongo-Driver Transaction Example - UpdateEmployeeInfo line 1742](https://github.com/mongodb/mongo-go-driver/blob/master/examples/documentation_examples/examples.go)
* models/db - user_service.go - Create is not multi request safe.  
See TODO note
* Document REST Api methods. Consider using [Swagger](https://towardsdatascience.com/setting-up-swagger-docs-for-golang-api-8d0442263641)

# Prerequisite

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


# Starting the Activity Server

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

# HTTP Usage

The following is a quick overview of running some of the current available commands. A more detail 
# Project Structure

This project is laid out as a Go module. 

The code is primarily laid out in a hierachy to support the notion of Model-View-Container. 

* models/client - Interfaces for returning data to the client
* models/db - Interfaces for reading and writing to the
* perm - Basic permissions for operations. The current setup is limited in scope and not implemented
    throught the project. It will be replaced by something more appropriate
* views - Interfaces to be used by the web server


# References
* [Make yourself a Go web server with MongoDb](https://medium.com/hackernoon/make-yourself-a-go-web-server-with-mongodb-go-on-go-on-go-on-48f394f24e)
* [Implementing JWT based authentication in Golang](https://www.sohamkamani.com/blog/golang/2019-01-01-jwt-authentication/)

