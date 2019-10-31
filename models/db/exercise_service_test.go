package db_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var TestExerciseFilename = "exercise_test.json"

// clearExercise drop all entries from the Exercise collection
func clearExercise(t *testing.T, ex *db.ExerciseService) {
	// Drop the Exercise collection table
	collection := ex.Connection.Database.Collection(ex.Config.CollectionName)
	assert.Nil(t, collection.Drop(context.TODO()))
}

// setup Setup the database for testing by creating a connection to the
// database and returning a handle to the ExerciseService. If desired
// via the clear flag the current Exercise collection entires can be
// dropped. Setting the load flag causes the predefined Exercise collection
// entires in TestExerciseFilename to be inserted into the Exercise collection.
func setupExercise(t *testing.T, clear bool, load bool) *db.ExerciseService {

	config := db.Config{Database: TestDatabase}
	ex, err := db.NewExerciseService(&config)
	assert.Nil(t, err)
	if clear {
		clearExercise(t, ex)
	}
	if load {
		err = ex.LoadFromFile(TestExerciseFilename)
		assert.Nil(t, err)
	}
	return ex
}

// teardown - perform database teardown to ensure each
// that the database is clean
func teardownExercise(t *testing.T, ex *db.ExerciseService) {
	clearExercise(t, ex)
}

func TestNewExerciseService(t *testing.T) {
	// Ensure that we got a handle to the service
	config := db.Config{Database: TestDatabase}
	ref, err := db.NewExerciseService(&config)
	assert.Nil(t, err)
	assert.NotNil(t, ref)
}

func TestCreateExercise(t *testing.T) {
	service := setupExercise(t, true, false)
	defer teardownExercise(t, service)

	// Add a new Exercise
	exercise := client.Exercise{
		Name:        "sit-ups",
		Description: "Description for sit-ups",
	}
	err := service.Create(&exercise)
	assert.Nil(t, err)
}

// TestCreateDuplicatExercise ensure an attempt to add a duplicate exercise fails
func TestCreateDuplicateExercise(t *testing.T) {
	service := setupExercise(t, true, false)
	defer teardownExercise(t, service)

	exercise := client.Exercise{
		Name:        "sit-ups",
		Description: "Description for sit-ups",
	}
	err := service.Create(&exercise)
	assert.Nil(t, err)

	// Attempt to add same exercise
	err = service.Create(&exercise)
	fmt.Println(err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestDeleteExercise(t *testing.T) {
	service := setupExercise(t, true, true)
	defer teardownExercise(t, service)

	// Use ID value for Jumping Jack
	err := service.Delete("5dab53b371aab123354e5cab")
	assert.Nil(t, err)

	// Attempt to delete a non existent exercise
	err = service.Delete(primitive.NewObjectID().Hex())
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "record found")

	// Attempt delete with bad ID
	err = service.Delete("Jumping Jack")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

func TestGetExerciseByID(t *testing.T) {
	service := setupExercise(t, true, true)
	defer teardownExercise(t, service)

	// Attempt to retrieve the an exercise from the populated Exercise collection
	e, err := service.GetByID("5dab53b371aab123354e5cab")
	assert.Equal(t, "Jumping Jack", e.Name)
	assert.Nil(t, err)

	// Attempt to retrieve a non existent user
	e, err = service.GetByID("noexist")
	assert.NotNil(t, err)
}

func TestGetExerciseByName(t *testing.T) {
	service := setupExercise(t, true, true)
	defer teardownExercise(t, service)

	// Attempt to retrieve the an exercise from the populated Exercise collection
	e, err := service.GetByName("Jumping Jack")
	assert.Equal(t, "Jumping Jack", e.Name)
	assert.Nil(t, err)

	// Attempt to retrieve a non existent user
	e, err = service.GetByName("noexist")
	assert.NotNil(t, err)
}

func TestGetAllExercises(t *testing.T) {
	service := setupExercise(t, true, true)
	defer teardownExercise(t, service)

	exercises, err := service.GetAll()
	assert.Nil(t, err)

	// The current json exercises collection contains 11 objects
	assert.Equal(t, 11, len(exercises))
}

func TestUpdateExercise(t *testing.T) {
	service := setupExercise(t, true, true)
	defer teardownExercise(t, service)

	// Get the record to modify
	e, err := service.GetByName("running")
	assert.Equal(t, "running", e.Name)
	assert.Nil(t, err)

	e.Description = "the action or movement of a runner."
	err = service.Update(e)
	assert.Nil(t, err)

	// Refetch the object and make sure it's been updated
	update, err := service.GetByName("running")
	assert.Nil(t, err)
	assert.Equal(t, e.Description, update.Description)
}

func TestUpdateNonExistent(t *testing.T) {
	service := setupExercise(t, false, false)

	exercise := client.Exercise{
		ID:   primitive.NewObjectID().Hex(),
		Name: "something",
	}
	err := service.Update(&exercise)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no match found")
}
