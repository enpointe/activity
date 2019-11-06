
# Activity Server - Version 2.0 

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
* Auditing - Understand the best method for recording audit level changes


# Issues Tracking 

[![Open issues](https://img.shields.io/github/issues/enpointe/activity)](https://github.com/enpointe/activity) [![Closed issues](https://img.shields.io/github/issues-closed/enpointe/activity)](https://github.com/enpointe/activity/issues?q=is%3Aissue+is%3Aclosed) [![Open PRs](https://img.shields.io/github/issues-pr/enpointe/activity)](https://github.com/enpointe/activity/pulls) [![Closed PRs](https://img.shields.io/github/issues-pr-closed/enpointe/activity)](https://github.com/enpointe/activity/pulls?q=is%3Apr+is%3Aclosed)

Some aspects of this project are being tracked via the GitHub Issues. Others are simply marked at various points in the code with TODO or tracked in Work Outstand. Gradually all these issues will be moved to Issues Tracking as the project formalizes a bit more.

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

The following is a quick overview of running some of the current available commands. 

## {serverURL}/login - Login in a User

The following shows the login process for a user. To log in a user the credential information for the user needs to be sent.
Upon successfully logging in a cookie 'token' is returned that represents the authentication token to use for subsequent requests.

```
$ curl -i -c activity.cookies -d '{"username":"admin", "password":"changeMe"}' -H "Content-Type: application/json" -X POST http://localhost:8080/login
HTTP/1.1 200 OK
Content-Type: application/json
Set-Cookie: token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6IjVkYzJlZTVhNTY3ODU1ZGUyMWYxMDcwYSIsIlVzZXJuYW1lIjoiYWRtaW4iLCJQcml2aWxlZ2UiOjIsImV4cCI6MTU3MzA1OTY4N30.N0Wdplcr2b10FUliqdqA_fhqSdtaoGb7Lfw8-w4X6N4; Max-Age=1200000
Date: Wed, 06 Nov 2019 16:41:27 GMT
Content-Length: 0
$ cat activity.cookies
# Netscape HTTP Cookie File
# https://curl.haxx.se/docs/http-cookies.html
# This file was generated by libcurl! Edit at your own risk.

localhost      	FALSE  	/      	FALSE  	1574258487     	token  	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6IjVkYzJlZTVhNTY3ODU1ZGUyMWYxMDcwYSIsIlVzZXJuYW1lIjoiYWRtaW4iLCJQcml2aWxlZ2UiOjIsImV4cCI6MTU3MzA1OTY4N30.N0Wdplcr2b10FUliqdqA_fhqSdtaoGb7Lfw8-w4X6N4
```

activity.cookies is where the JWT token 'token' is stored with contains the authorization which will be used for subsequent commands to the activity server.

## {serverURL}/logout - Log out a user

The logout operation will log the user out of the current session and clear the token cookie.  Using curl this
requires us to use both the -b and -c options.

```
$ curl -b activity.cookies -c activity.cookies  -H "Content-Type: application/json" -X GET http://localhost:8080/logout
$ cat activity.cookies
# Netscape HTTP Cookie File
# https://curl.haxx.se/docs/http-cookies.html
# This file was generated by libcurl! Edit at your own risk.

localhost      	FALSE  	/      	FALSE  	1572926089     	token
```

**Notice:** that the logout of the user causes the cookie token to expire

## {serverURL}/users/ - Retrieve informationa about all known users

```
$ curl -i -b activity.cookies -c activity.cookies  -H "Content-Type: application/json" -X GET http://localhost:8080/users/
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 06 Nov 2019 16:43:18 GMT
Content-Length: 178

[{"_id":{"$oid":"5db8e02b0e7aa732afd7fbc1"},"user_id":"customer1","password":"$2a$10$wN/f8n8RGoHbngeO3Ub1L.wCZxZhb98xu1mk9Ysvct47.YggtDrr.","privilege":0},
{"_id":{"$oid":"5db8e02b0e7aa732afd7fbc2"},"user_id":"staff","password":"$2a$10$u05g/b0wHvZAmEwd6ABt2.gPPQqD.LlypM.fta6coleTpQ0qdgRdK","privilege":1},
{"_id":{"$oid":"5db8e02b0e7aa732afd7fbc4"},"user_id":"admin","password":"$2a$10$JwIOnVsJ1kFrcAZ657R0Euid19Ybapys7AtWfCVAqbJTDMx3oYnEu","privilege":2}]
```

## {serverURL}/user/{id} - Retrieve information about a particular user

Retrieve information about a particular user

```
$ curl -i -b activity.cookies -c activity.cookies  -H "Content-Type: application/json" -X GET http://localhost:8080/user/5dc08c9d989368d8f439e39a
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 06 Nov 2019 16:43:50 GMT
Content-Length: 88

{"id":"5dc08c9d989368d8f439e39a","username":"admin","password":"-","privilege":"admin"}
```

## {serverURL}/user/create - Create a new user

```
$ curl -i -b activity.cookies -d '{"username":"kitty", "password":"1Me0w4u@H", "privilege": "staff"}' -H "Content-Type: application/json" -X POST http://localhost:8080/user/create
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 06 Nov 2019 16:44:37 GMT
Content-Length: 34

{"id":"5dc2f8751e1d7704072214b8"}
```

## {serverURL}/user/delete/{id} - Delete user with id

```
$ curl -i -b activity.cookies -c activity.cookies  -H "Content-Type: application/json" -X POST http://localhost:8080/user/delete/5dc3227a3ff84c7a8374616d
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 06 Nov 2019 19:44:11 GMT
Content-Length: 21

{"deletedEntries":1}
```

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

