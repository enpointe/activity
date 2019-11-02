package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/perm"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UsersCollection name of the collection used to hold user information
const UsersCollection = "users"

// UserService holds a entry to the User Collection in the database
type UserService struct {
	Config     *Config
	context    context.Context
	Connection *Connection
	Collection *mongo.Collection
}

// NewUserService create a new instance of the User Service
// An optional optParam may be specified
// optParam[0] = DatabaseName
func NewUserService(ctx context.Context, config *Config) (*UserService, error) {
	connection, err := GetConnection(config)
	if err != nil {
		return nil, err
	}
	if len(config.CollectionName) == 0 {
		config.CollectionName = UsersCollection
	}
	collection := connection.Database.Collection(config.CollectionName)
	return &UserService{config, ctx, connection, collection}, nil
}

// Create add a new user to the database
func (s *UserService) Create(user *client.User) error {
	u, err := NewUser(user)
	if err != nil {
		return err
	}

	// Check to make sure a user with the specified user ID doesn't already exist
	cursor := s.Collection.FindOne(s.context, bson.M{
		"user_id": user.Username,
	})
	if err = cursor.Err(); err == nil {
		// A match for that user already exists
		err = fmt.Errorf("A entry matching the userID '%s' already exists", user.Username)
		log.Print(err)
		return err
	}

	_, err = s.Collection.InsertOne(s.context, &u)
	if err != nil {
		err = fmt.Errorf("Unable to store user data in database, %s", err)
		log.Print(err)
	}

	return err
}

// DeleteUserData deletes the user associated with id and all
// associated information associated with that user. Once deleted
// the information can not be recovered.
func (s *UserService) DeleteUserData(hexid string) error {
	// TODO Must add in deletes for the users exercise logs when
	// add to project. This must be done as a transaction to ensure that
	// we don't end up with a partial deletion of data

	idPrimitive, err := primitive.ObjectIDFromHex(hexid)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", hexid, err)
		return err
	}
	results, err := s.Collection.DeleteOne(s.context, bson.M{"_id": idPrimitive})
	if err != nil {

		err = fmt.Errorf("failed to delete %s, %s", hexid, err)
		log.Print(err)
	}
	if results.DeletedCount == 0 {
		err = fmt.Errorf("failed to delete %s, no entry for record found", hexid)
	}
	return err
}

// DeleteAll deletes all user records
func (s *UserService) DeleteAll() error {
	return s.Collection.Drop(s.context)
}

// findOne internal method for retrieving the user record
func (s *UserService) findOne(filter interface{}) (*User, error) {
	cursor := s.Collection.FindOne(s.context, filter)
	if err := cursor.Err(); err != nil {
		log.Printf("user not found, %s", err)
		err = fmt.Errorf("user not found")
		return nil, err
	}
	var user User
	cursor.Decode(&user)
	return &user, nil
}

// AdminUserExists basic test to ensure at a minimum one
// admin account exists
func (s *UserService) AdminUserExists() (bool, error) {
	filter := bson.M{"privilege": perm.Admin}
	user, err := s.findOne(filter)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

// GetByID retrieve the user record via the passed in username
func (s *UserService) GetByID(id string) (*client.User, error) {
	idPrimitive, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", id, err)
		return nil, err
	}
	filter := bson.M{"_id": idPrimitive}
	user, err := s.findOne(filter)
	if err != nil {
		return nil, err
	}
	cUser := user.Convert()
	return &cUser, nil
}

// GetByUsername retrieve the user record via the passed in username
func (s *UserService) GetByUsername(username string) (*client.User, error) {
	filter := bson.D{primitive.E{Key: "user_id", Value: username}}
	user, err := s.findOne(filter)
	if err != nil {
		return nil, err
	}
	cUser := user.Convert()
	return &cUser, nil
}

// GetAll return information about all users
func (s *UserService) GetAll() ([]*client.User, error) {
	var results []*client.User

	// Check to make sure a user with the specified user ID doesn't already exist
	cursor, err := s.Collection.Find(s.context, bson.D{{}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(s.context)

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cursor.Next(s.context) {

		// create a value into which the single document can be decoded
		var elem User
		err := cursor.Decode(&elem)
		if err != nil {
			log.Printf("GetAll: failed to decode %s", elem)
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

// Update update the user record represented by u.ID.
// Only the Username, Password, and Privilege fields may be updated
func (s *UserService) Update(u *client.User) error {
	idPrimitive, err := primitive.ObjectIDFromHex(u.ID)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", u.ID, err)
		return err
	}
	filter := bson.M{"_id": idPrimitive}
	update := bson.M{"$set": bson.M{
		"username":  u.Username,
		"password":  u.Password,
		"privilege": perm.Convert(u.Privilege),
	}}
	updateResult, err := s.Collection.UpdateOne(s.context, filter, update)
	if err != nil {
		err = fmt.Errorf("failed to update user %s, %s", u.ID, err)
		return err
	}
	if updateResult.ModifiedCount != 1 {
		err = fmt.Errorf("failed to update user %s, no match found", u.ID)
		return err
	}
	return nil
}

// Validate validate the credentials of the user
func (s *UserService) Validate(c *client.Credentials) (*client.User, error) {

	filter := bson.D{primitive.E{Key: "user_id", Value: c.Username}}
	user, err := s.findOne(filter)
	if err != nil {
		return nil, fmt.Errorf("invalid username/password")
	}

	err = user.comparePassword(c.Password)
	if err != nil {
		return nil, fmt.Errorf("invalid username/password")
	}
	cUser := user.Convert()
	return &cUser, nil
}

// LoadFromFile load json data from a file directly into a database.
// If the ID field of the user data is not set, ie ObjectID.IsZero(),
// a new ObjectID will be created for the user. The form of the json
// file is compatible with mongoexport --type json --jsonArray
func (s *UserService) LoadFromFile(filename string) error {
	// Load values from JSON file to model
	byteValues, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var users []User
	err = json.Unmarshal(byteValues, &users)
	if err != nil {
		return err
	}
	log.Print(users)
	var userToAdd []interface{}
	for _, u := range users {
		// Check to see if the ID field is set. If not set it
		if u.ID.IsZero() {
			// Not set
			u.ID = primitive.NewObjectID()
		}
		userToAdd = append(userToAdd, u)
	}
	// Insert users into DB
	ordered := false
	insertOptions := &options.InsertManyOptions{Ordered: &ordered}
	_, err = s.Collection.InsertMany(s.context, userToAdd, insertOptions)
	return err
}
