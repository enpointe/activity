package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/enpointe/activity/models/client"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ExerciseCollection name of the collection used to hold exercise information
const ExerciseCollection = "testdata/exercises"

// ExerciseService holds a entry to the Exercise Collection in the database
type ExerciseService struct {
	Collection *mongo.Collection
	log        *log.Logger
}

// NewExerciseService create a new instance of the Exercise Service
func NewExerciseService(database *mongo.Database, logger *log.Logger) (*ExerciseService, error) {
	collection := database.Collection(ExerciseCollection)
	return &ExerciseService{
		Collection: collection, log: logger}, nil
}

// Create adds a new exercise to the database
func (s *ExerciseService) Create(ctx context.Context, ex *client.Exercise) error {
	if len(strings.TrimSpace(ex.Name)) == 0 {
		err := fmt.Errorf("exercise name must be specified")
		return err
	}
	exercise := NewExercise(ex)

	// Check to make sure a exercise with the specified exercise name doesn't already exist
	cursor := s.Collection.FindOne(ctx, bson.M{"name": exercise.Name})
	if err := cursor.Err(); err == nil {
		// A match for that user already exists
		err = fmt.Errorf("A entry matching the exercise name '%s' already exists", exercise.Name)
		log.Debug(err)
		return err
	}

	_, err := s.Collection.InsertOne(ctx, &exercise)
	if err != nil {
		log.Errorf("Insert of %s failed, %s", exercise.Name, err)
	}
	return err
}

// Delete remove the exercise with the specified id from the database
func (s *ExerciseService) Delete(ctx context.Context, hexid string) error {

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

// DeleteAll deletes all exercise records
func (s *ExerciseService) DeleteAll(ctx context.Context) error {
	return s.Collection.Drop(context.TODO())
}

// Update update an existing exercise. Only the name and/or description
// field can be updated
func (s *ExerciseService) Update(ctx context.Context, e *client.Exercise) error {
	idPrimitive, err := primitive.ObjectIDFromHex(e.ID)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", e.ID, err)
		return err
	}
	filter := bson.M{"_id": idPrimitive}
	update := bson.M{"$set": bson.M{"name": e.Name, "description": e.Description}}
	updateResult, err := s.Collection.UpdateOne(ctx, filter, update)
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

func (s *ExerciseService) getOne(ctx context.Context, filter interface{}) (*Exercise, error) {
	var exercise Exercise
	cursor := s.Collection.FindOne(ctx, filter)
	if err := cursor.Err(); err != nil {
		err = fmt.Errorf("exercise not found, %s", err)
		return nil, err
	}
	err := cursor.Decode(&exercise)
	if err != nil {
		log.Errorf("failed to decode exercise %s", err)
		return nil, err
	}
	return &exercise, nil
}

// GetByID retrieve the details of an exercise
func (s *ExerciseService) GetByID(ctx context.Context, hexid string) (*client.Exercise, error) {
	idPrimitive, err := primitive.ObjectIDFromHex(hexid)
	if err != nil {
		err = fmt.Errorf("invalid id %s, %s", hexid, err)
		return nil, err
	}
	exercise, err := s.getOne(ctx, bson.M{"_id": idPrimitive})
	if err != nil {
		return nil, err
	}
	cExercise := exercise.Convert()
	return &cExercise, nil
}

// GetByName retrieve the details of an exercise
func (s *ExerciseService) GetByName(ctx context.Context, name string) (*client.Exercise, error) {
	exercise, err := s.getOne(ctx, bson.D{primitive.E{Key: "name", Value: name}})
	if err != nil {
		return nil, err
	}
	cExercise := exercise.Convert()
	return &cExercise, nil
}

// GetAll retrieve a list of all known exercises
func (s *ExerciseService) GetAll(ctx context.Context) ([]*client.Exercise, error) {
	var results []*client.Exercise
	// Check to make sure a user with the specified user ID doesn't already exist
	cursor, err := s.Collection.Find(ctx, bson.D{{}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cursor.Next(ctx) {

		// create a value into which the single document can be decoded
		var elem Exercise
		err := cursor.Decode(&elem)
		if err != nil {
			log.Errorf("failed to decode exercise %s", elem)
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
func (s *ExerciseService) LoadFromFile(ctx context.Context, filename string) error {
	// Load values from JSON file to model
	byteValues, err := ioutil.ReadFile(filename)
	if err != nil {
		log.WithFields(log.Fields{
			"filename":           filename,
			"string(byteValues)": string(byteValues),
		}).Debug(err)
		return err
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
	_, err = s.Collection.InsertMany(ctx, exercisesToAdd, insertOptions)
	return err
}
