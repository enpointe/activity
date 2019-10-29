package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/enpointe/activity/models/client"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UsersCollection name of the collection used to hold user information
const UsersCollection = "users"

// UserService holds a entry to the User Collection in the database
type UserService struct {
	config     *Config
	collection *mongo.Collection
}

// NewUserService create a new instance of the User Service
// An optional optParam may be specified
// optParam[0] = DatabaseName
func NewUserService(config *Config) (*UserService, error) {
	connection, err := GetConnection(config)
	if err != nil {
		return nil, err
	}
	if len(config.CollectionName) == 0 {
		config.CollectionName = UsersCollection
	}
	collection := connection.Database.Collection(config.CollectionName)
	return &UserService{config, collection}, nil
}

// CreateUser add a new user to the database
func (s *UserService) CreateUser(user *client.User) error {
	u, err := NewUser(user)
	if err != nil {
		return err
	}

	// Set how long to wait for operation to complete before timing out
	context, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	// Check to make sure a user with the specified user ID doesn't already exist
	cursor := s.collection.FindOne(context, bson.M{
		"user_id": user.Username,
	})
	if err = cursor.Err(); err == nil {
		// A match for that user already exists
		err = fmt.Errorf("A entry matching the userID '%s' already exists", user.Username)
		log.Print(err)
		return err
	}

	_, err = s.collection.InsertOne(context, &u)
	if err != nil {
		log.Printf("Insert of %s failed, %s", user, err)
	}

	return err
}

// DeleteAllUserData deletes a user and all information associated with that user
// from the database.
func (s *UserService) DeleteAllUserData(user *client.User) error {
	// Set how long to wait for operation to complete before timing out
	context, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	// TODO Must add in deletes for the users exercise logs when
	// add to project
	// TODO This must be done as a transaction to ensure that
	// we don't end up with a partial deletion of data

	idDoc := bson.D{primitive.E{Key: "user_id", Value: user.Username}}
	_, err := s.collection.DeleteOne(context, idDoc)
	if err != nil {
		log.Printf("Delete of %s failed, %s", user, err)
	}
	return err
}

// getUserByUsername internal method for retrieving the user record
func (s *UserService) getUserByUsername(username string) (User, error) {
	var user User
	context, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	idDoc := bson.D{primitive.E{Key: "user_id", Value: username}}
	cursor := s.collection.FindOne(context, idDoc)
	if err := cursor.Err(); err != nil {
		log.Printf("user '%s' not found in collection %s, %s", username, UsersCollection, err)
		err = fmt.Errorf("user with specified username not found: %s", username)
		return user, err
	}
	cursor.Decode(&user)
	return user, nil
}

// GetUserByUsername retrieve the user record via the passed in username
func (s *UserService) GetUserByUsername(username string) (client.User, error) {
	var hUser client.User
	user, err := s.getUserByUsername(username)
	if err != nil {
		return hUser, err
	}
	hUser = client.User{
		ID:       user.ID.Hex(),
		Username: user.Username,
		Password: "-",
	}
	return hUser, nil
}

// GetAllUsers return information about all users
func (s *UserService) GetAllUsers() ([]*client.User, error) {

	// Set how long to wait for operation to complete before timing out
	context, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	var results []*client.User

	// Check to make sure a user with the specified user ID doesn't already exist
	cursor, err := s.collection.Find(context, bson.D{{}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context)

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cursor.Next(context) {

		// create a value into which the single document can be decoded
		var elem User
		err := cursor.Decode(&elem)
		if err != nil {
			log.Printf("GetAllUsers: failed to decode %s", elem)
			return nil, err
		}

		user := client.User{
			ID:        elem.ID.Hex(),
			Username:  elem.Username,
			Password:  "-",
			Privilege: elem.Privilege.String(),
		}

		results = append(results, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// Validate validate the credentials of the user
func (s *UserService) Validate(c *client.Credentials) (client.User, error) {
	var hUser client.User
	user, err := s.getUserByUsername(c.Username)
	if err != nil {
		return client.User{}, fmt.Errorf("invalid username/password")
	}

	err = user.comparePassword(c.Password)
	if err != nil {
		return client.User{}, fmt.Errorf("invalid username/password")
	}
	hUser = client.User{
		ID:       user.ID.Hex(),
		Username: user.Username,
		Password: "-",
	}
	return hUser, nil
}
