
This is a simplistic implemenation of golang activity module that provides a simple in memory implementation for recording activities and the time spent on each activity via JSON requests.

# Installation 

As this is a Go module the code cannot be installed directly inside the $GOPATH workarea.

For information on using GO modules see the article 
[Create projects independent of $GOPATH using Go Modules](https://medium.com/mindorks/create-projects-independent-of-gopath-using-go-modules-802260cdfb51)

# Bugs

* When used in the context of an imported module ClearStats() is coming up as undefined. AddAction() and GetStats() are being properly linked.

# Possible Future Enhancements

NOTE: As indicated this is an overly simplist implemenation. Possible future enchancements might include

* Add persistent store to store action data and time over restarts
* Add methods to restrict actions to a well defined set of actions
* Define a clearly identifiable time interval
* Add ability to auto record time period via start/stop action methods
