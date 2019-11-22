package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/enpointe/activity/controllers"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo/options"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/enpointe/activity/docs"
)

// catchFatal - log any Fatal error conditions before exiting.
func catchFatal() {
	err := recover()
	if err != nil {
		entry := err.(*logrus.Entry)
		log.WithFields(logrus.Fields{
			"err_level":   entry.Level,
			"err_message": entry.Message,
		}).Error("Server Panic")
	}
}

// swagger Serve swagger
func swagger(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	httpSwagger.WrapHandler.ServeHTTP(w, r)
}

// @title Activity API
// @version 2.1
// @description This is a activity server, used to record workouts.

// @license.name Mit License
// @license.url https://github.com/enpointe/activity/blob/master/LICENSE

// @host localhost:8080
// @BasePath /
// @tokenUrl https://jwt.io/
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	defer catchFatal()
	dbURI := flag.String("dbURI", "mongodb://localhost:27017",
		"URI used to connect to the mongo database, default is mongodb://localhost:27017")
	adminPasswd := flag.String("admin", "", "Create an admin user and assign it the specified password")
	logLevel := flag.String(
		"level", "warn", "The logging level to use (error, warn, info, debug, trace)")
	flag.Parse()

	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		fmt.Printf("unsupported log level specified '%s', valid values are: error, warn, info, debug, trace", *logLevel)
		os.Exit(-1)
	}
	if level == log.PanicLevel || level == log.FatalLevel {
		fmt.Println("Ignoring requested log level, setting to", log.ErrorLevel)
		level = log.ErrorLevel
	}

	clientOptions := options.Client().ApplyURI(*dbURI)
	sOptions := []controllers.ServerOption{controllers.DBOptions(clientOptions)}
	if len(*adminPasswd) > 0 {
		sOptions = append(sOptions, controllers.CreateAdminUser([]byte(*adminPasswd)))
	}

	var filename string = "activity.log"
	// Create the log file if doesn't exist. And append to it if it already exists.
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	Formatter := new(log.TextFormatter)
	// You can change the Timestamp format. But you have to use the same date and time.
	// "2006-02-02 15:04:06" Works. If you change any digit, it won't work
	// ie "Mon Jan 2 15:04:05 MST 2006" is the reference time. You can't change it
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)
	if err != nil {
		// Cannot open log file. Logging to stderr
		fmt.Println(err)
	} else {
		log.SetOutput(f)
		fmt.Println("Logging output to :", filename)
	}
	log.SetLevel(level)
	log.Debug("Starting HTTP Server")
	server, err := controllers.NewServerService(false, sOptions...)
	if err != nil {
		fmt.Printf("%s\n\n", err.Error())
		os.Exit(-2)
	}
	router := httprouter.New()
	router.POST("/login", server.Login)
	router.POST("/logout", server.Logout)
	router.POST("/users", server.CreateUser)
	router.DELETE("/users", server.DeleteUser)
	router.GET("/users", server.GetUsers)
	router.GET("/users/:id", server.GetUser)
	router.PATCH("/users/", server.UpdateUserPassword)

	// programatically set swagger info
	docs.SwaggerInfo.Title = "Activity API"
	docs.SwaggerInfo.Description = "This is a Activity JSON server."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	router.GET("/swagger/", swagger)
	log.Fatal(http.ListenAndServe(":8080", router))
}
