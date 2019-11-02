package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/enpointe/activity/models/client"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ExerciseCollection name of the collection used to hold exercise information
const ExerciseCollection = "testdata/exercises"

// ExerciseService holds a entry to the Exercise Collection in the database
type ExerciseService struct {
	Config     *Config
	Connection *Connection
	Collection *mongo.Collection
}

// NewExerciseService create a new instance of the Exercise Service
func NewExerciseService(config *Config) (*ExerciseService, error) {
	connection, err := GetConnection(config)
	if err != nil {
		return nil, err
	}
	if len(config.CollectionName) == 0 {
		config.CollectionName = ExerciseCollection
	}
	collection := connection.Database.Collection(config.CollectionName)
	return &ExerciseService{config, connection, collection}, nil
}

// Create adds a new exercise to the database
func (s *ExerciseService) Create(ex *client.Exercise) error {
	if len(strings.TrimSpace(ex.Name)) == 0 {
		err := fmt.Errorf("exercise name must be specified")
		return err
	}
	exercise := NewExercise(ex)

	// Set how long to wait for operation to complete before timing out
	context, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	// Check to make sure a exercise with the specified exercise name doesn't already exist
	cursor := s.Collection.FindOne(context, bson.M{"name": exercise.Name})
	if err := cursor.Err(); err == nil {
		// A match for that user already exists
		err = fmt.Errorf("A entry matching the exercise name '%s' already exists", exercise.Name)
		log.Print(err)
		return err
	}

	_, err := s.Collection.InsertOne(context, &exercise)
	if err != nil {
		log.Printf("Insert of %s failed, %s", exercise.Name, err)
	}
	return err
}

// Delete remove the exercise with the specified id from the database
func (s *ExerciseService) Delete(hexid string) error {
	// Set how long to wait for operation to complete before timing out
	context, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	idPrimitive, err := primitive.ObjectIDFromHex(hexid)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", hexid, err)
		return err
	}
	results, err := s.Collection.DeleteOne(context, bson.M{"_id": idPrimitive})
	if err != nil {

		err = fmt.Errorf("failed to delete %s, %s", hexid, err)
		log.Print(err)
	}
	if results.DeletedCount == 0 {
		err = fmt.Errorf("failed to delete %s, no entry for record found", hexid)
	}
	return err
}

// DeleteAll deletes all exercise records
func (s *ExerciseService) DeleteAll() error {
	return s.Collection.Drop(context.TODO())
}

// Update update an existing exercise. Only the name and/or description
// field can be updated
func (s *ExerciseService) Update(e *client.Exercise) error {
	context, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	idPrimitive, err := primitive.ObjectIDFromHex(e.ID)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", e.ID, err)
		return err
	}
	filter := bson.M{"_id": idPrimitive}
	update := bson.M{"$set": bson.M{"name": e.Name, "description": e.Description}}
	updateResult, err := s.Collection.UpdateOne(context, filter, update)
	if err != nil {
		err = fmt.Errorf("failed to update exercise %s, %s", e.ID, err)
		return err
	}
	if updateResult.ModifiedCount != 1 {
		err = fmt.Errorf("failed to update exercise %s, no match found", e.ID)
		return err
	}
	return nil
}

func (s *ExerciseService) getOne(filter interface{}) (*Exercise, error) {
	var exercise Exercise
	context, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	cursor := s.Collection.FindOne(context, filter)
	if err := cursor.Err(); err != nil {
		err = fmt.Errorf("exercise not found, %s", err)
		return nil, err
	}
	err := cursor.Decode(&exercise)
	if err != nil {
		log.Printf("failed to decode exercise %s", err)
		return nil, err
	}
	return &exercise, nil
}

// GetByID retrieve the details of an exercise
func (s *ExerciseService) GetByID(hexid string) (*client.Exercise, error) {
	idPrimitive, err := primitive.ObjectIDFromHex(hexid)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", hexid, err)
		return nil, err
	}
	exercise, err := s.getOne(bson.M{"_id": idPrimitive})
	if err != nil {
		return nil, err
	}
	cExercise := exercise.Convert()
	return &cExercise, nil
}

// GetByName retrieve the details of an exercise
func (s *ExerciseService) GetByName(name string) (*client.Exercise, error) {
	exercise, err := s.getOne(bson.D{primitive.E{Key: "name", Value: name}})
	if err != nil {
		return nil, err
	}
	cExercise := exercise.Convert()
	return &cExercise, nil
}

// GetAll retrieve a list of all known exercises
func (s *ExerciseService) GetAll() ([]*client.Exercise, error) {

	// Set how long to wait for operation to complete before timing out
	context, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	var results []*client.Exercise
	// Check to make sure a user with the specified user ID doesn't already exist
	cursor, err := s.Collection.Find(context, bson.D{{}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context)

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cursor.Next(context) {

		// create a value into which the single document can be decoded
		var elem Exercise
		err := cursor.Decode(&elem)
		if err != nil {
			log.Printf("failed to decode exercise %s", elem)
			return nil, err
		}

		exercise := client.Exercise{
			ID:          elem.ID.Hex(),
			Name:        elem.Name,
			Description: elem.Description,
		}
		results = append(results, &exercise)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// LoadFromFile load json data from a file directly into a database.
// If the ID field of the exercise data is not set, ie ObjectID.IsZero(),
// a new ObjectID will be created for the exercise.
func (s *ExerciseService) LoadFromFile(filename string) error {
	// Load values from JSON file to model
	byteValues, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	var ex []Exercise
	json.Unmarshal(byteValues, &ex)
	var exercisesToAdd []interface{}
	for _, e := range ex {
		// Check to see if the ID field is set. If not set it
		if e.ID.IsZero() {
			// Not set
			e.ID = primitive.NewObjectID()
		}
		exercisesToAdd = append(exercisesToAdd, e)
	}
	// Insert exercise into DB
	ordered := false
	insertOptions := &options.InsertManyOptions{Ordered: &ordered}
	_, err = s.Collection.InsertMany(context.Background(), exercisesToAdd, insertOptions)
	return err
}
