package db

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connection handle to the mongo client and database
type Connection struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// DefaultDatabase name of the database
const DefaultDatabase = "activity"

//GetConnection returns connection to the specified mongo database
func GetConnection(config *Config) (*Connection, error) {
	if len(config.URL) == 0 {
		config.URL = "mongodb://localhost:27017"
	}
	var connection Connection
	clientOptions := options.Client().ApplyURI(config.URL)
	var err error
	connection.Client, err = mongo.NewClient(clientOptions)
	if err != nil {
		err = fmt.Errorf("failed to open a connection to MongoDB, %s", err)
		log.Print(err)
		return nil, err
	}
	err = connection.Client.Connect(context.TODO())
	if err != nil {
		log.Print(err)
		return nil, err
	}
	if len(config.Database) == 0 {
		config.Database = DefaultDatabase
	}
	connection.Database = connection.Client.Database(config.Database)
	return &connection, nil
}
