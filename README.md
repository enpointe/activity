
# Version 2.0 (In development)

This is a rework/expansion of the version 1.0 implementation of this activity logger which was written as a take home project for a job interview.

This project is primarily intended as a learning exercise to learn how to program in Golang.

In order to make the origional project a bit more interesting the overall project is being expanded to include

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
* Initial privilege support added to user structure. 

# Work outstanding

* Improve unit test
    * Investigate how to mock database calls
    * Modify view such that we can use test database instead of actual database
* Integration Level tests
* Add more robust logging via [logrus](https://github.com/sirupsen/logrus])
* Use of perm for controlling which github.com/enpointe/activity/view methods can be called is too simplistic. Consider alternatives like https://github.com/casbin/casbin
* Create custom error types so that more realistic https status code can be returned when a error occurs at the database level due to bad data in request
* Add database configuration for security
* Add cli interface that allows initial admin user to be created
* Add mechanism for prepopulating database with Exercises
* Examine whether view code should have some context cancel in it. See [Stack Overflow question](https://stackoverflow.com/questions/47179024/how-to-check-if-a-request-was-cancelled)
* JWT secret key needs to be configurable.
    * Consider creating a cli interface for this
* Need to create an initial admin user in order for http interfaces to function.
    * Consider creating a cli interface for this
* perm authorization is a bit comberson. Consider adding a simple RBAC authorization on methods or 
    * [Casbin](https://github.com/casbin/casbin)
    * [goRBAC](https://github.com/mikespook/gorbac)

# Prerequisite

As this tools uses MongoDB it will be necessary to install and start the mongod server. The following schema opperations will need to be executed
in order to ensure things are setup properly

```
$ mongod
```

Start the server.  The server needs to be accessible via mongodb://localhost:27017

```
$ mongod
```


As all the components for this project have not been completed it will be necessary to prestub the database initially
This can be done via 


# Project Structure

This project is laid out as a Go module. As such the code cannot be installed directly inside the $GOPATH workarea. 

The code is primarily laid out in a hierachy to support the notion of Model-View-Container. 


# References
* [Make yourself a Go web server with MongoDb](https://medium.com/hackernoon/make-yourself-a-go-web-server-with-mongodb-go-on-go-on-go-on-48f394f24e)
* [Implementing JWT based authentication in Golang](https://www.sohamkamani.com/blog/golang/2019-01-01-jwt-authentication/)

