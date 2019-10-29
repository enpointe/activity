package db_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/stretchr/testify/assert"
)

var TestDatabase = "testActivity"

func clearCollection(t *testing.T, config *db.Config) {
	connection, err := db.GetConnection(config)
	if err == nil {
		// Drop the user collection table
		collection := connection.Database.Collection(db.UsersCollection)
		assert.Nil(t, collection.Drop(context.TODO()))
	}
}

func TestNewUserService(t *testing.T) {
	// Ensure that we got a handle to the service
	config := db.Config{Database: TestDatabase}
	ref, err := db.NewUserService(&config)
	assert.Nil(t, err)
	assert.NotNil(t, ref)
}

func TestCreateUser(t *testing.T) {
	config := db.Config{Database: TestDatabase}
	// Clear the test collection at the start and end of the test
	clearCollection(t, &config)
	defer func() {
		clearCollection(t, &config)
	}()
	userService, err := db.NewUserService(&config)
	assert.Nil(t, err)

	// Add a new user
	user := client.User{
		Username: "wwomen",
		Password: "tellTheTruth",
	}
	err = userService.CreateUser(&user)
	assert.Nil(t, err)

	// Attempt to add same user
	err = userService.CreateUser(&user)
	fmt.Println(err)
	assert.Contains(t, err.Error(), "already exists")

	// Attempt to add a different user
	user = client.User{
		Username: "root",
		Password: "changeMe",
	}
	err = userService.CreateUser(&user)
	assert.Nil(t, err)
}

func TestValidate(t *testing.T) {
	config := db.Config{Database: TestDatabase}
	// Clear the test collection at the start and end of the test
	clearCollection(t, &config)
	defer func() {
		clearCollection(t, &config)
	}()
	userService, err := db.NewUserService(&config)
	assert.Nil(t, err)

	// Add a new user
	user := client.User{
		Username: "root",
		Password: "changeme",
	}
	err = userService.CreateUser(&user)
	assert.Nil(t, err)

	// Attempt to login in with the proper credentials
	credentials := client.Credentials{
		Username: user.Username,
		Password: user.Password,
	}
	clientUser, err := userService.Validate(&credentials)
	assert.Nil(t, err)
	assert.True(t, len(clientUser.ID) != 0)
	assert.Equal(t, user.Username, clientUser.Username)
	assert.Equal(t, clientUser.Password, "-")

	// Atempt to login with the improper credentials
	credentials.Password = "wrongPassword"
	clientUser, err = userService.Validate(&credentials)
}

func TestGetUserByName(t *testing.T) {
	config := db.Config{Database: TestDatabase}
	// Clear the test collection at the start and end of the test
	clearCollection(t, &config)
	defer func() {
		clearCollection(t, &config)
	}()

	userService, err := db.NewUserService(&config)
	assert.Nil(t, err)

	// Add a test user
	testUser := client.User{
		Username: "testUser",
		Password: "changeme",
	}
	err = userService.CreateUser(&testUser)
	assert.Nil(t, err)

	// Attempt to retrieve the testUser
	retUser, err := userService.GetUserByUsername(testUser.Username)
	assert.Equal(t, retUser.Username, testUser.Username)
	assert.Nil(t, err)

	// Attempt to retrieve non existent user
	retUser, err = userService.GetUserByUsername("nonExistentUser")
	assert.Contains(t, err.Error(), "not found")
}

func TestGetAllUsers(t *testing.T) {
	config := db.Config{Database: TestDatabase}
	// Clear the test collection at the start and end of the test
	clearCollection(t, &config)
	defer func() {
		clearCollection(t, &config)
	}()

	userService, err := db.NewUserService(&config)
	assert.Nil(t, err)

	wwomen := client.User{
		Username: "wwomen",
		Password: "tellTheTruth",
	}
	err = userService.CreateUser(&wwomen)
	assert.Nil(t, err)

	admin := client.User{
		Username: "admin",
		Password: "changeMe",
	}
	err = userService.CreateUser(&admin)
	assert.Nil(t, err)

	// Add a test user
	ironman := client.User{
		Username: "ironman",
		Password: "tonystark",
	}
	err = userService.CreateUser(&ironman)
	assert.Nil(t, err)

	// Attempt to retrieve the testUser
	users, err := userService.GetAllUsers()
	assert.Nil(t, err)
	assert.Equal(t, len(users), 3)
}
