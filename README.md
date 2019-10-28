
Version 2 In Progress

This is a rework/expansion of the version 1 implementation of this activity logger.

This version is being expanded to:

* Add backend storage using MongoDB
* Add a front end http interface to access the JSON methods
* Add ability to store information about the user exercising
* Add ability to store information about the exercise being perform
* Add ability to record the time spent performing a specific exercise for a specific user
* Add basic security to the REST api calls to ensure proper access



# Development

Local development machines need to have following tools installed and working properly:

- [Docker](https:://www.docker.com) for running a full-time containerized development environment.

Windows users need to additionally have an Unix-shell emulator to be able to run utility scripts (Git Bash is recommended).

To start a docker container with the full running development environement execute the following command after cloning a copy of this development tree

<code>
$ scripts/scripts/start-dev-container.sh
</code>

# Issues Tracking 

[![Open issues](https://img.shields.io/github/issues/enpointe/activity)](https://github.com/enpointe/activity) [![Closed issues](https://img.shields.io/github/issues-closed/enpointe/activity)](https://github.com/enpointe/activity/issues?q=is%3Aissue+is%3Aclosed) [![Open PRs](https://img.shields.io/github/issues-pr/enpointe/activity)](https://github.com/enpointe/activity/pulls) [![Closed PRs](https://img.shields.io/github/issues-pr-closed/enpointe/activity)](https://github.com/enpointe/activity/pulls?q=is%3Apr+is%3Aclosed)

This project use Github project, issues and pull requests to manage and track issues.

# Project Structure

This project is laid out as a Go module. As such the code cannot be installed directly inside the $GOPATH workarea. 

The docker container script "start-dev-container.sh" has been setup to ensure a consistent development environment.

# References
* [Make yourself a Go web server with MongoDb(https://medium.com/hackernoon/make-yourself-a-go-web-server-with-mongodb-go-on-go-on-go-on-48f394f24e)]
* [Context keys in Go(https://medium.com/@matryer/context-keys-in-go-5312346a868d#.hb4spbx1a)]
* [Implementing OAuth 2.0 with Go(https://www.sohamkamani.com/blog/golang/2018-06-24-oauth-with-golang/)]
# TODO

* Add more robust logging via [logrus(https://github.com/sirupsen/logrus])]
* Create custom error types so that more realistic https status code can be returned when 
a error occurs at the database level due to bad data in request
* Add security to http request
* Add configuration to the database
