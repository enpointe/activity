package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/perm"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// UsersCollection name of the collection used to hold user information
const UsersCollection = "users"

// UserService holds a entry to the User Collection in the database
type UserService struct {
	Collection *mongo.Collection
	client     *mongo.Client
}

// NewUserService create a new instance of the User Service
func NewUserService(database *mongo.Database) (*UserService, error) {
	// Set majority write concern
	//wMajority := writeconcern.New(writeconcern.WMajority())
	//collectionOptions := &options.CollectionOptions{WriteConcern: wMajority}
	//collection := database.Collection(UsersCollection, collectionOptions)
	collection := database.Collection(UsersCollection)

	return &UserService{
		Collection: collection}, nil
}

// Create add a new user to the database
func (s *UserService) Create(ctx context.Context, user *client.UserCreate) (string, error) {
	u, err := NewUser(user)
	if err != nil {
		return "", err
	}
	filter := bson.M{"user_id": user.Username}
	// session, err := s.client.StartSession()
	// if err != nil {
	// 	return "", err
	// }
	// if err = session.StartTransaction(); err != nil {
	// 	return "", err
	// }
	// var resultID string
	// err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

	// 	cursor := s.Collection.FindOne(sc, filter)
	// 	if err = cursor.Err(); err == nil {
	// 		// A match for that user already exists
	// 		err = fmt.Errorf("A entry matching the userID '%s' already exists", user.Username)
	// 		log.Debug(err)
	// 		return err
	// 	}

	// 	result, err := s.Collection.InsertOne(sc, &u)
	// 	if err != nil {
	// 		err = fmt.Errorf("Unable to store user data in database, %s", err)
	// 		log.Error(err)
	// 		return err
	// 	}
	// 	resultID = result.InsertedID.(primitive.ObjectID).Hex()
	// 	err = session.CommitTransaction(sc)
	// 	return err
	// })
	// session.EndSession(ctx)
	// return resultID, nil

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
		log.WithFields(log.Fields{
			"document": u,
		}).Debugf("user collection InsertOne() failed: %s", err)
		err = fmt.Errorf("Unable to store user data in database, %s", err)
		log.Error(err)
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), err
}

// DeleteUserData deletes the user associated with id and all
// associated information associated with that user. Once deleted
// the information can not be recovered.
// Return delete count if successful, error otherwise
func (s *UserService) DeleteUserData(ctx context.Context, hexid string) (int, error) {
	idPrimitive, err := primitive.ObjectIDFromHex(hexid)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", hexid, err)
		return 0, err
	}

	// session, err := s.client.StartSession()
	// if err != nil {
	// 	return err
	// }
	// if err = session.StartTransaction(); err != nil {
	// 	return err
	// }
	// // Create a transaction for deleting all user records.
	// // Currently this is only the user record but in the future will
	// // include the user exercise log entries
	// err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

	// 	results, err := s.Collection.DeleteOne(ctx, bson.M{"_id": idPrimitive})
	// 	if err != nil {

	// 		err = fmt.Errorf("failed to delete %s, %s", hexid, err)
	// 		log.Error(err)
	// 	}
	// 	if results.DeletedCount == 0 {
	// 		err = fmt.Errorf("failed to delete %s, no entry for record found", hexid)
	// 	}

	// 	err = session.CommitTransaction(sc)
	// 	return err
	// })
	// session.EndSession(ctx)results, err := s.Collection.DeleteOne(ctx, bson.M{"_id": idPrimitive})
	filter := bson.M{"_id": idPrimitive}
	results, err := s.Collection.DeleteOne(ctx, filter)
	if err != nil {
		log.WithFields(log.Fields{
			"filter": filter,
		}).Debugf("user collection DeleteOne() failed: %s", err)
		err = fmt.Errorf("failed to delete %s, %s", hexid, err)
		log.Error(err)
	}
	if results.DeletedCount == 0 {
		log.Infof("failed to delete %s, no entry for record found", hexid)
	}
	return int(results.DeletedCount), err
}

// DeleteAll deletes all user records
func (s *UserService) DeleteAll(ctx context.Context) error {
	return s.Collection.Drop(ctx)
}

// findOne internal method for retrieving the user record
func (s *UserService) findOne(ctx context.Context, filter interface{}) (*User, error) {
	cursor := s.Collection.FindOne(ctx, filter)
	if err := cursor.Err(); err != nil {
		log.WithFields(log.Fields{
			"filter": filter,
		}).Debugf("user collection FindOne() failed: %s", err)
		err = fmt.Errorf("user not found")
		return nil, err
	}
	var user User
	cursor.Decode(&user)
	return &user, nil
}

// AdminUserExists basic test to ensure at a minimum one
// admin account exists
func (s *UserService) AdminUserExists(ctx context.Context) bool {
	filter := bson.M{"privilege": perm.Admin}
	user, err := s.findOne(ctx, filter)
	if err != nil {
		log.Debug(err.Error())
		return false
	}
	return user != nil
}

// GetByID retrieve the user record via the passed in username
func (s *UserService) GetByID(ctx context.Context, id string) (*client.UserInfo, error) {
	idPrimitive, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", id, err)
		log.Debug(err)
		return nil, err
	}
	filter := bson.M{"_id": idPrimitive}
	user, err := s.findOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	cUser := user.Convert()
	return &cUser, nil
}

// GetByUsername retrieve the user record via the passed in username
func (s *UserService) GetByUsername(ctx context.Context, username string) (*client.UserInfo, error) {
	filter := bson.D{primitive.E{Key: "user_id", Value: username}}
	user, err := s.findOne(ctx, filter)
	if err != nil {
		log.WithFields(log.Fields{
			"filter": filter,
		}).Debugf("user collection findOne() failed: %s", err)
		return nil, err
	}
	cUser := user.Convert()
	return &cUser, nil
}

// GetAll return information about all users. Password
// information for each user will simply be returned as "-"
func (s *UserService) GetAll(ctx context.Context) ([]*client.UserInfo, error) {
	var results []*client.UserInfo

	// Set the projection for the request to not
	// return the password field.
	projection := bson.D{primitive.E{Key: "password", Value: 0}}
	cursor, err := s.Collection.Find(ctx,
		bson.D{{}},
		options.Find().SetProjection(projection))
	if err != nil {
		log.WithFields(log.Fields{
			"projection": projection,
		}).Debugf("user collection Find() failed: %s", err)
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

		user := client.UserInfo{
			ID:        elem.ID.Hex(),
			Username:  elem.Username,
			Privilege: elem.Privilege.String(),
		}

		results = append(results, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *UserService) update(ctx context.Context, filter bson.M, update bson.M) (int64, error) {
	updateResult, err := s.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.WithFields(log.Fields{
			"filter": filter,
			"update": update,
		}).Debugf("user collection UpdateOne()) failed: %s", err)
		return 0, err
	}
	if updateResult.ModifiedCount != 1 {
		err = errors.New("no match found")
		return 0, err
	}
	return updateResult.ModifiedCount, nil
}

// Update update the user record represented by u.ID.
// Only the Username, Password, and Privilege fields may be updated.
// The password is assumed to have been encrptyed for storage.
func (s *UserService) Update(ctx context.Context, u *client.UserUpdate) (int, error) {
	idPrimitive, err := primitive.ObjectIDFromHex(u.ID)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", u.ID, err)
		return 0, err
	}
	filter := bson.M{"_id": idPrimitive}
	update := bson.M{"$set": bson.M{
		"user_id":   u.Username,
		"password":  u.Password,
		"privilege": perm.Convert(u.Privilege),
	}}
	cnt, err := s.update(ctx, filter, update)
	if err != nil {
		err = fmt.Errorf("Failed to update user '%s', %s", u.ID, err)
		return 0, err
	}
	return int(cnt), nil
}

// UpdatePassword updates the password for the specified ID.
// The password is assumed to be encrypted.
func (s *UserService) UpdatePassword(ctx context.Context, passInfo *client.PasswordUpdate) (int, error) {
	idPrimitive, err := primitive.ObjectIDFromHex(passInfo.ID)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", passInfo.ID, err)
		return 0, err
	}
	filter := bson.M{"_id": idPrimitive}
	update := bson.M{"$set": bson.M{
		"password": passInfo.NewPassword,
	}}
	cnt, err := s.update(ctx, filter, update)
	if err != nil {
		err = fmt.Errorf("Failed to update user '%s', %s", passInfo.ID, err)
		return 0, err
	}
	return int(cnt), nil
}

// Validate validate the credentials of the user
func (s *UserService) Validate(ctx context.Context, c *client.Credentials) (*client.UserInfo, error) {

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
