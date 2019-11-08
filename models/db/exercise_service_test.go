package db_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var testExerciseFilename = "testdata/exercise_test.json"

// setup Setup the database for testing by creating a connection to the
// database and returning a handle to the ExerciseService. If desired
// via the clear flag the current Exercise collection entires can be
// dropped. Setting the load flag causes the predefined Exercise collection
// entires in TestExerciseFilename to be inserted into the Exercise collection.
func SetupExercise(t *testing.T, clear bool, load bool) *db.ExerciseService {
	clientOptions := options.Client().ApplyURI(testDatabaseURL)
	ctx := context.TODO()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	assert.NoError(t, err)

	database := client.Database(testDatabase)
	ex, err := db.NewExerciseService(database, log.StandardLogger())
	assert.NoError(t, err)
	if clear {
		err = ex.DeleteAll(ctx)
		assert.NoError(t, err)
	}
	if load {
		err = ex.LoadFromFile(ctx, testExerciseFilename)
		assert.NoError(t, err, "Load of json data from %s failed", testExerciseFilename)
	}
	return ex
}

// teardown - perform database teardown to ensure each
// that the database is clean
func TeardownExercise(t *testing.T, ex *db.ExerciseService) {
	err := ex.DeleteAll(context.TODO())
	assert.NoError(t, err)
}

func TestCreateExercise(t *testing.T) {
	service := SetupExercise(t, true, false)
	defer TeardownExercise(t, service)
	ctx := context.TODO()

	// Add a new Exercise
	exercise := client.Exercise{
		Name:        "sit-ups",
		Description: "Description for sit-ups",
	}
	err := service.Create(ctx, &exercise)
	assert.NoError(t, err)
}

func TestCreateExerciseNoName(t *testing.T) {
	service := SetupExercise(t, true, false)
	defer TeardownExercise(t, service)
	ctx := context.TODO()

	// Add a new Exercise with no name specified
	exercise := client.Exercise{
		Description: "Description for sit-ups",
	}
	err := service.Create(ctx, &exercise)
	assert.Error(t, err)
}

// TestCreateDuplicatExercise ensure an attempt to add a duplicate exercise fails
func TestCreateDuplicateExercise(t *testing.T) {
	service := SetupExercise(t, true, false)
	defer TeardownExercise(t, service)
	ctx := context.TODO()

	exercise := client.Exercise{
		Name:        "sit-ups",
		Description: "Description for sit-ups",
	}
	err := service.Create(ctx, &exercise)
	assert.NoError(t, err)

	// Attempt to add same exercise
	err = service.Create(ctx, &exercise)
	fmt.Println(err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestDeleteExercise(t *testing.T) {
	service := SetupExercise(t, true, true)
	defer TeardownExercise(t, service)
	ctx := context.TODO()

	// Use ID value for Jumping Jack
	err := service.Delete(ctx, "5dab53b371aab123354e5cab")
	assert.NoError(t, err)

	// Attempt to delete a non existent exercise
	err = service.Delete(ctx, primitive.NewObjectID().Hex())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record found")

	// Attempt delete with bad ID
	err = service.Delete(ctx, "Jumping Jack")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

func TestGetExerciseByID(t *testing.T) {
	service := SetupExercise(t, true, true)
	defer TeardownExercise(t, service)
	ctx := context.TODO()

	// Attempt to retrieve the an exercise from the populated Exercise collection
	e, err := service.GetByID(ctx, "5dab53b371aab123354e5cab")
	assert.Equal(t, "Jumping Jack", e.Name)
	assert.NoError(t, err)

	// Attempt to retrieve a non existent user
	e, err = service.GetByID(ctx, "noexist")
	assert.Error(t, err)
}

func TestGetExerciseByName(t *testing.T) {
	service := SetupExercise(t, true, true)
	defer TeardownExercise(t, service)
	ctx := context.TODO()

	// Attempt to retrieve the an exercise from the populated Exercise collection
	e, err := service.GetByName(ctx, "Jumping Jack")
	assert.Equal(t, "Jumping Jack", e.Name)
	assert.NoError(t, err)

	// Attempt to retrieve a non existent user
	e, err = service.GetByName(ctx, "noexist")
	assert.Error(t, err)
}

func TestGetAllExercises(t *testing.T) {
	service := SetupExercise(t, true, true)
	defer TeardownExercise(t, service)
	ctx := context.TODO()

	exercises, err := service.GetAll(ctx)
	assert.NoError(t, err)

	// The current json exercises collection contains 11 objects
	assert.Equal(t, 11, len(exercises))
}

func TestUpdateExercise(t *testing.T) {
	service := SetupExercise(t, true, true)
	defer TeardownExercise(t, service)
	ctx := context.TODO()

	// Get the record to modify
	e, err := service.GetByName(ctx, "running")
	assert.Equal(t, "running", e.Name)
	assert.NoError(t, err)

	e.Description = "the action or movement of a runner."
	err = service.Update(ctx, e)
	assert.NoError(t, err)

	// Refetch the object and make sure it's been updated
	update, err := service.GetByName(ctx, "running")
	assert.NoError(t, err)
	assert.Equal(t, e.Description, update.Description)

	// Attempt update with bad ID
	e.ID = "Jumping Jack"
	err = service.Update(ctx, e)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

func TestUpdateNonExistent(t *testing.T) {
	service := SetupExercise(t, false, false)
	ctx := context.TODO()

	exercise := client.Exercise{
		ID:   primitive.NewObjectID().Hex(),
		Name: "something",
	}
	err := service.Update(ctx, &exercise)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no match found")
}

func TestLoadExFromFile(t *testing.T) {
	// Test load from non existant file
	service := SetupExercise(t, false, false)
	ctx := context.TODO()
	err := service.LoadFromFile(ctx, "someFile")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")

	// Load in a exercise without a ID specified. Check to make sure
	// an ID is generated
	err = service.LoadFromFile(ctx, "testdata/ex_load_test.json")
	assert.NoError(t, err)
	if err != nil {
		ex, err := service.GetByName(ctx, "Jumping Jack")
		assert.NoError(t, err)
		assert.NotNil(t, ex)
		if ex != nil {
			assert.NotNil(t, ex.ID)
		}
	}

	// Attempt to load in a badly formed json file
	err = service.LoadFromFile(ctx, "testdata/invalid.json")
	assert.Error(t, err)
}
