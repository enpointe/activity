package db

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/enpointe/activity/models/client"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ExerciseCollection name of the collection used to hold exercise information
const ExerciseCollection = "exercises"

// ExerciseService holds a entry to the Exercise Collection in the database
type ExerciseService struct {
	collection *mongo.Collection
}

// NewExerciseService create a new instance of the Exercise Service
// An optional optParam may be specified
// optParam[0] = DatabaseName
func NewExerciseService(config *Config) (*ExerciseService, error) {
	connection, err := GetConnection(config)
	if err != nil {
		return nil, err
	}
	if len(config.CollectionName) == 0 {
		config.CollectionName = ExerciseCollection
	}
	collection := connection.Database.Collection(config.CollectionName)
	return &ExerciseService{collection}, nil
}

// CreateExercise adds a new exercise to the database
func (s *ExerciseService) CreateExercise(ex *client.Exercise) error {
	if len(strings.TrimSpace(ex.Name)) != 0 {
		err := fmt.Errorf("exercise name must be specified")
		return err
	}
	exercise := NewExercise(ex)

	// Set how long to wait for operation to complete before timing out
	context, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	// Check to make sure a user with the specified user ID doesn't already exist
	cursor := s.collection.FindOne(context, bson.M{"name": exercise.Name})
	if err := cursor.Err(); err == nil {
		// A match for that user already exists
		err = fmt.Errorf("A entry matching the exercise name '%s' already exists", exercise.Name)
		log.Print(err)
		return err
	}

	_, err := s.collection.InsertOne(context, &exercise)
	if err != nil {
		log.Printf("Insert of %s failed, %s", exercise.Name, err)
	}
	return err
}

// DeleteExercise delete and existing exercise
func (s *ExerciseService) DeleteExercise(id string) error {

	// Set how long to wait for operation to complete before timing out
	context, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	idDoc := bson.D{primitive.E{Key: "_id", Value: id}}
	_, err := s.collection.DeleteOne(context, idDoc)
	if err != nil {
		log.Printf("Delete of %s failed, %s", id, err)
	}
	return err
}

// UpdateExercise update and existing exercise
func (s *ExerciseService) UpdateExercise(e *Exercise) (*Exercise, error) {
	return nil, nil
}

// GetExercise retrieve the details of an exercise
func (s *ExerciseService) GetExercise(id string) (Exercise, error) {
	var exercise Exercise
	context, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	idDoc := bson.D{primitive.E{Key: "_id", Value: id}}
	cursor := s.collection.FindOne(context, idDoc)
	if err := cursor.Err(); err != nil {
		log.Printf("exercise '%s' not found in collection %s, %s", id, ExerciseCollection, err)
		err = fmt.Errorf("exercise with specified id not found: %s", id)
		return exercise, err
	}
	cursor.Decode(&exercise)
	return exercise, nil
}

// GetAllExercises retrieve a list of all known exercises
func (s *ExerciseService) GetAllExercises() ([]Exercise, error) {
	return nil, nil
}

// GetAllExerciseNames retrieve a list of the names of all known exercises
func (s *ExerciseService) GetAllExerciseNames() ([]string, error) {
	return nil, nil
}
