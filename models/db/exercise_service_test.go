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

var TestExerciseFilename = "testdata/exercise_test.json"

// ClearExercise drop all entries from the Exercise collection
func clearExercise(t *testing.T, ex *db.ExerciseService) {
	// Drop the Exercise collection table
	collection := ex.Connection.Database.Collection(db.ExerciseCollection)
	assert.Nil(t, collection.Drop(context.TODO()))
}

// setup Setup the database for testing by creating a connection to the
// database and returning a handle to the ExerciseService. If desired
// via the clear flag the current Exercise collection entires can be
// dropped. Setting the load flag causes the predefined Exercise collection
// entires in TestExerciseFilename to be inserted into the Exercise collection.
func SetupExercise(t *testing.T, clear bool, load bool) *db.ExerciseService {

	config := db.Config{Database: testDatabase}
	ex, err := db.NewExerciseService(&config)
	assert.NoError(t, err)
	if clear {
		err := ex.DeleteAll()
		assert.NoError(t, err)
	}
	if load {
		err = ex.LoadFromFile(TestExerciseFilename)
		assert.NoError(t, err)
	}
	return ex
}

// teardown - perform database teardown to ensure each
// that the database is clean
func TeardownExercise(t *testing.T, ex *db.ExerciseService) {
	clearExercise(t, ex)
}

func TestNewExerciseService(t *testing.T) {
	// Ensure that we got a handle to the service
	config := db.Config{Database: testDatabase}
	ref, err := db.NewExerciseService(&config)
	assert.NoError(t, err)
	assert.NotNil(t, ref)
}

func TestCreateExercise(t *testing.T) {
	service := SetupExercise(t, true, false)
	defer TeardownExercise(t, service)

	// Add a new Exercise
	exercise := client.Exercise{
		Name:        "sit-ups",
		Description: "Description for sit-ups",
	}
	err := service.Create(&exercise)
	assert.NoError(t, err)
}

// TestCreateDuplicatExercise ensure an attempt to add a duplicate exercise fails
func TestCreateDuplicateExercise(t *testing.T) {
	service := SetupExercise(t, true, false)
	defer TeardownExercise(t, service)

	exercise := client.Exercise{
		Name:        "sit-ups",
		Description: "Description for sit-ups",
	}
	err := service.Create(&exercise)
	assert.NoError(t, err)

	// Attempt to add same exercise
	err = service.Create(&exercise)
	fmt.Println(err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestDeleteExercise(t *testing.T) {
	service := SetupExercise(t, true, true)
	defer TeardownExercise(t, service)

	// Use ID value for Jumping Jack
	err := service.Delete("5dab53b371aab123354e5cab")
	assert.NoError(t, err)

	// Attempt to delete a non existent exercise
	err = service.Delete(primitive.NewObjectID().Hex())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record found")

	// Attempt delete with bad ID
	err = service.Delete("Jumping Jack")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

func TestGetExerciseByID(t *testing.T) {
	service := SetupExercise(t, true, true)
	defer TeardownExercise(t, service)

	// Attempt to retrieve the an exercise from the populated Exercise collection
	e, err := service.GetByID("5dab53b371aab123354e5cab")
	assert.Equal(t, "Jumping Jack", e.Name)
	assert.NoError(t, err)

	// Attempt to retrieve a non existent user
	e, err = service.GetByID("noexist")
	assert.Error(t, err)
}

func TestGetExerciseByName(t *testing.T) {
	service := SetupExercise(t, true, true)
	defer TeardownExercise(t, service)

	// Attempt to retrieve the an exercise from the populated Exercise collection
	e, err := service.GetByName("Jumping Jack")
	assert.Equal(t, "Jumping Jack", e.Name)
	assert.NoError(t, err)

	// Attempt to retrieve a non existent user
	e, err = service.GetByName("noexist")
	assert.Error(t, err)
}

func TestGetAllExercises(t *testing.T) {
	service := SetupExercise(t, true, true)
	defer TeardownExercise(t, service)

	exercises, err := service.GetAll()
	assert.NoError(t, err)

	// The current json exercises collection contains 11 objects
	assert.Equal(t, 11, len(exercises))
}

func TestUpdateExercise(t *testing.T) {
	service := SetupExercise(t, true, true)
	defer TeardownExercise(t, service)

	// Get the record to modify
	e, err := service.GetByName("running")
	assert.Equal(t, "running", e.Name)
	assert.NoError(t, err)

	e.Description = "the action or movement of a runner."
	err = service.Update(e)
	assert.NoError(t, err)

	// Refetch the object and make sure it's been updated
	update, err := service.GetByName("running")
	assert.NoError(t, err)
	assert.Equal(t, e.Description, update.Description)
}

func TestUpdateNonExistent(t *testing.T) {
	service := SetupExercise(t, false, false)

	exercise := client.Exercise{
		ID:   primitive.NewObjectID().Hex(),
		Name: "something",
	}
	err := service.Update(&exercise)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no match found")
}
