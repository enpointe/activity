package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/perm"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UsersCollection name of the collection used to hold user information
const UsersCollection = "users"

// UserService holds a entry to the User Collection in the database
type UserService struct {
	Collection *mongo.Collection
}

// NewUserService create a new instance of the User Service
func NewUserService(database *mongo.Database) (*UserService, error) {
	collection := database.Collection(UsersCollection)
	return &UserService{
		Collection: collection}, nil
}

// Create add a new user to the database
func (s *UserService) Create(ctx context.Context, user *client.User) (string, error) {
	u, err := NewUser(user)
	if err != nil {
		return "", err
	}

	filter := bson.M{"user_id": user.Username}

	// TODO - This is not safe as it is possible for a user_id to be created
	// between our FindOne and the InsertOne Request. It appears we want
	// do a upsert/$setOnInsert using FindOneUpdate.

	// Check to make sure a user with the specified user ID doesn't already exist
	cursor := s.Collection.FindOne(ctx, filter)
	if err = cursor.Err(); err == nil {
		// A match for that user already exists
		err = fmt.Errorf("A entry matching the userID '%s' already exists", user.Username)
		log.Debug(err)
		return "", err
	}

	result, err := s.Collection.InsertOne(ctx, &u)
	if err != nil {
		err = fmt.Errorf("Unable to store user data in database, %s", err)
		log.Error(err)
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), err
}

// DeleteUserData deletes the user associated with id and all
// associated information associated with that user. Once deleted
// the information can not be recovered.
func (s *UserService) DeleteUserData(ctx context.Context, hexid string) error {
	// TODO Must add in deletes for the users exercise logs when
	// add to project. This must be done as a transaction to ensure that
	// we don't end up with a partial deletion of data
	//

	idPrimitive, err := primitive.ObjectIDFromHex(hexid)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", hexid, err)
		return err
	}
	results, err := s.Collection.DeleteOne(ctx, bson.M{"_id": idPrimitive})
	if err != nil {

		err = fmt.Errorf("failed to delete %s, %s", hexid, err)
		log.Error(err)
	}
	if results.DeletedCount == 0 {
		err = fmt.Errorf("failed to delete %s, no entry for record found", hexid)
	}
	return err
}

// DeleteAll deletes all user records
func (s *UserService) DeleteAll(ctx context.Context) error {
	return s.Collection.Drop(ctx)
}

// findOne internal method for retrieving the user record
func (s *UserService) findOne(ctx context.Context, filter interface{}) (*User, error) {
	cursor := s.Collection.FindOne(ctx, filter)
	if err := cursor.Err(); err != nil {
		log.Debugf("user not found, %s", err)
		err = fmt.Errorf("user not found")
		return nil, err
	}
	var user User
	cursor.Decode(&user)
	return &user, nil
}

// AdminUserExists basic test to ensure at a minimum one
// admin account exists
func (s *UserService) AdminUserExists(ctx context.Context) (bool, error) {
	filter := bson.M{"privilege": perm.Admin}
	user, err := s.findOne(ctx, filter)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

// GetByID retrieve the user record via the passed in username
func (s *UserService) GetByID(ctx context.Context, id string) (*client.User, error) {
	idPrimitive, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", id, err)
		log.Debug(err)
		return nil, err
	}
	filter := bson.M{"_id": idPrimitive}
	user, err := s.findOne(ctx, filter)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	cUser := user.Convert()
	return &cUser, nil
}

// GetByUsername retrieve the user record via the passed in username
func (s *UserService) GetByUsername(ctx context.Context, username string) (*client.User, error) {
	log.WithFields(log.Fields{
		"username": username,
	}).Debug("GetByUserName - Enter")
	filter := bson.D{primitive.E{Key: "user_id", Value: username}}
	user, err := s.findOne(ctx, filter)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	cUser := user.Convert()
	return &cUser, nil
}

// GetAll return information about all users. Password
// information for each user will simply be returned as "-"
func (s *UserService) GetAll(ctx context.Context) ([]*client.User, error) {
	log.Debug("GetAll - Enter")
	var results []*client.User

	// Set the projection for the request to not
	// return the password field.
	projection := bson.D{primitive.E{Key: "password", Value: 0}}
	cursor, err := s.Collection.Find(ctx,
		bson.D{{}},
		options.Find().SetProjection(projection))
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cursor.Next(ctx) {

		// create a value into which the single document can be decoded
		var elem User
		err := cursor.Decode(&elem)
		if err != nil {
			log.Errorf("failed to decode %s", elem)
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
func (s *UserService) Update(ctx context.Context, u *client.User) error {
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
	updateResult, err := s.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		err = fmt.Errorf("failed to update user %s, %s", u.ID, err)
		log.Error(err)
		return err
	}
	if updateResult.ModifiedCount != 1 {
		err = fmt.Errorf("failed to update user %s, no match found", u.ID)
		return err
	}
	return nil
}

// Validate validate the credentials of the user
func (s *UserService) Validate(ctx context.Context, c *client.Credentials) (*client.User, error) {

	filter := bson.D{primitive.E{Key: "user_id", Value: c.Username}}
	user, err := s.findOne(ctx, filter)
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
func (s *UserService) LoadFromFile(ctx context.Context, filename string) error {
	// Load values from JSON file to model
	byteValues, err := ioutil.ReadFile(filename)
	if err != nil {
		err := fmt.Errorf("failed to read file %s, %s", filename, err)
		log.Debug(err)
		return err
	}
	var users []User
	err = json.Unmarshal(byteValues, &users)
	if err != nil {
		log.WithFields(log.Fields{
			"filename":           filename,
			"string(byteValues)": string(byteValues),
		}).Debug(err)
		return err
	}
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
	_, err = s.Collection.InsertMany(ctx, userToAdd, insertOptions)
	return err
}
