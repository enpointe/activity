package db_test

import (
	"context"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"gotest.tools/assert"
)

var TestDatabase = "testActivity"

func clearCollection(t *testing.T, config *db.Config) {
	connection, err := db.GetConnection(config)
	if err == nil {
		// Drop the user collection table
		collection := connection.Database.Collection(db.UsersCollection)
		assert.NilError(t, collection.Drop(context.TODO()))
	}
}

func TestNewUserService(t *testing.T) {
	// Ensure that we got a handle to the service
	config := db.Config{
		Database: TestDatabase,
	}
	ref, err := db.NewUserService(&config)
	assert.NilError(t, err)
	assert.Assert(t, ref != nil)
}

func TestCreateUser(t *testing.T) {
	config := db.Config{Database: TestDatabase}
	// Clear the test collection at the start and end of the test
	clearCollection(t, &config)
	defer func() {
		clearCollection(t, &config)
	}()
	userService, err := db.NewUserService(&config)
	assert.NilError(t, err)

	// Add a new user
	user := client.User{
		Username: "wwomen",
		Password: "tellTheTruth",
	}
	err = userService.CreateUser(&user)
	assert.NilError(t, err)

	// Attempt to add same user
	err = userService.CreateUser(&user)
	assert.ErrorContains(t, err, "already exists")

	// Attempt to add a different user
	user = client.User{
		Username: "root",
		Password: "changeMe",
	}
	err = userService.CreateUser(&user)
	assert.NilError(t, err)
}

func TestLogin(t *testing.T) {
	config := db.Config{Database: TestDatabase}
	// Clear the test collection at the start and end of the test
	clearCollection(t, &config)
	defer func() {
		clearCollection(t, &config)
	}()
	userService, err := db.NewUserService(&config)
	assert.NilError(t, err)

	// Add a new user
	user := client.User{
		Username: "root",
		Password: "changeme",
	}
	err = userService.CreateUser(&user)
	assert.NilError(t, err)

	// Attempt to login in with the proper credentials
	credentials := client.Credentials{
		Username: user.Username,
		Password: user.Password,
	}
	clientUser, err := userService.Login(&credentials)
	assert.NilError(t, err)
	assert.Assert(t, len(clientUser.ID) != 0)
	assert.Equal(t, user.Username, clientUser.Username)
	assert.Equal(t, clientUser.Password, "-")

	// Atempt to login with the improper credentials
	credentials.Password = "wrongPassword"
	clientUser, err = userService.Login(&credentials)
}

func TestGetUserByName(t *testing.T) {
	config := db.Config{Database: TestDatabase}
	// Clear the test collection at the start and end of the test
	clearCollection(t, &config)
	defer func() {
		clearCollection(t, &config)
	}()

	userService, err := db.NewUserService(&config)
	assert.NilError(t, err)

	// Add a test user
	testUser := client.User{
		Username: "testUser",
		Password: "changeme",
	}
	err = userService.CreateUser(&testUser)
	assert.NilError(t, err)

	// Attempt to retrieve the testUser
	retUser, err := userService.GetUserByUsername(testUser.Username)
	assert.Equal(t, retUser.Username, testUser.Username)
	assert.Assert(t, err == nil)

	// Attempt to retrieve non existent user
	retUser, err = userService.GetUserByUsername("nonExistentUser")
	assert.ErrorContains(t, err, "not found")
}
