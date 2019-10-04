
This is a simplistic implemenation of golang activity module that provides a simple in memory implementation for recording activities and the time spent on each activity via JSON requests. It is intended primary as the means to explore and learn about go lang.

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


# Possible Future Enhancements

NOTE: As indicated this is an overly simplist implemenation being used to explore and learn about golang. Possible future enchancements might include

* Add persistent store to store action data and time over restarts
* Add methods to restrict actions to a well defined set of actions
* Define a clearly identifiable time interval
* Add ability to auto record time period via start/stop action methods
