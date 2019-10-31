package db_test

import (
	"context"
	"testing"

	"github.com/enpointe/activity/perm"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/stretchr/testify/assert"
)

var TestDatabase = "testActivity"
var TestUserFilename = "user_test.json"

// clearUser drop all entries from the user collection
func clearUser(t *testing.T, us *db.UserService) {
	// Drop the user collection table
	collection := us.Connection.Database.Collection(us.Config.CollectionName)
	assert.Nil(t, collection.Drop(context.TODO()))
}

// setup Setup the database for testing by creating a connection to the
// database and returning a handle to the UserService. If desired
// via the clear flag the current user collection entires can be
// dropped. Setting the load flag causes the predefined user collection
// entires in TestUserFilename to be inserted into the user collection.
func setupUser(t *testing.T, clear bool, load bool) *db.UserService {

	config := db.Config{Database: TestDatabase}
	us, err := db.NewUserService(&config)
	assert.Nil(t, err)
	if clear {
		clearUser(t, us)
	}
	if load {
		err = us.LoadFromFile(TestUserFilename)
		assert.Nil(t, err, "Load of json data from %s failed", TestUserFilename)
	}
	return us
}

// teardown - perform database teardown to ensure each
// that the database is clean
func teardownUser(t *testing.T, us *db.UserService) {
	clearUser(t, us)
}

// TestNewUserService perform a simple instantiation test
func TestNewUserService(t *testing.T) {
	// Ensure that we got a handle to the service
	config := db.Config{Database: TestDatabase}
	ref, err := db.NewUserService(&config)
	assert.Nil(t, err)
	assert.NotNil(t, ref)
}

// TestCreateUser test the creation of users.
func TestCreateUser(t *testing.T) {
	userService := setupUser(t, true, false)
	defer teardownUser(t, userService)

	// Add a customer
	user := client.User{
		Username: "customer1",
		Password: "password",
	}
	err := userService.Create(&user)
	assert.Nil(t, err)

	// Add a new staff user
	user = client.User{
		Username:  "wwomen",
		Password:  "password",
		Privilege: perm.Staff.String(),
	}
	err = userService.Create(&user)
	assert.Nil(t, err)

	// Attempt to add a different user
	user = client.User{
		Username:  "admin",
		Password:  "changeMe",
		Privilege: perm.Admin.String(),
	}
	err = userService.Create(&user)
	assert.Nil(t, err)
}

// TestCreateDuplicatUser ensure an attempt to add a duplicate user fails
func TestCreateDuplicateUser(t *testing.T) {
	userService := setupUser(t, true, false)
	defer teardownUser(t, userService)

	// Add a customer
	user := client.User{
		Username: "customer1",
		Password: "password",
	}
	err := userService.Create(&user)
	assert.Nil(t, err)

	// Attempt to add same user
	err = userService.Create(&user)
	assert.Contains(t, err.Error(), "already exists")
}

// TestValidate ensure that a Validate correctly
// handles a login with correct and incorrect passwords
func TestValidate(t *testing.T) {
	setupUser(t, true, true)
	userService := setupUser(t, true, true)

	defer teardownUser(t, userService)

	// Attempt to login in with the proper credentials
	// NOTE: This test will fail if the credentials
	// in TestUserFilename are changed
	credentials := client.Credentials{
		Username: "customer1",
		Password: "password",
	}
	user, err := userService.Validate(&credentials)
	assert.Nil(t, err)
	assert.True(t, len(user.ID) != 0)
	assert.Equal(t, credentials.Username, user.Username)
	assert.Equal(t, "-", user.Password)

	// Atempt to login with the improper credentials
	credentials.Password = "wrongPassword"
	user, err = userService.Validate(&credentials)
	assert.NotNil(t, err)
}

func TestGetUserByName(t *testing.T) {
	userService := setupUser(t, true, true)
	defer teardownUser(t, userService)

	// Attempt to retrieve the user populated to the user collection
	retUser, err := userService.GetByUsername("admin")
	assert.Equal(t, "admin", retUser.Username)
	assert.Nil(t, err)

	retUser, err = userService.GetByUsername("staff")
	assert.Equal(t, "staff", retUser.Username)
	assert.Nil(t, err)

	retUser, err = userService.GetByUsername("customer1")
	assert.Equal(t, "customer1", retUser.Username)
	assert.Nil(t, err)

	// Attempt to retrieve a non existent user
	retUser, err = userService.GetByUsername("noexist")
	assert.NotNil(t, err)
}

func TestGetAllUsers(t *testing.T) {
	userService := setupUser(t, true, true)
	defer teardownUser(t, userService)

	// Attempt to retrieve the populated users
	users, err := userService.GetAll()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(users))
}

func TestDeleteUser(t *testing.T) {
	userService := setupUser(t, true, true)
	defer teardownUser(t, userService)

	// Attempt to delete an existing user from the database
	user, err := userService.GetByUsername("customer1")
	assert.Equal(t, "customer1", user.Username)
	assert.Nil(t, err)

	err = userService.DeleteUserData(user.ID)
	assert.Nil(t, err)

	// Attempt to delete a non existent user
	err = userService.DeleteUserData(primitive.NewObjectID().Hex())
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to delete")

	// Attempt delete with bad ID
	err = userService.DeleteUserData("4dddkalkdlajbeee")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}
