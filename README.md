
This is a simplistic implemenation of golang activity package that provides a simple in memory implementation for recording activities and the time spent on each activity via JSON requests.

# Installation 

Please use the standard `go get` command to build and install this module on Linux, OSX, and Windows

```
go get -u github.com/enpointe/activity
```

Also, if not already set, you have to add the $GOPATH\bin directory to your PATH variable.

# Possible Future Extentions

* Add persistent store to store action data and time over restarts
* Add public clear method for clearing statistics
* Add methods to restrict actions to a well defined set of actions
