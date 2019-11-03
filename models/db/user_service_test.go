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

const testDatabaseURL = "mongodb://localhost:27017"
const testDatabase string = "testActivity"
const testUserFilename string = "testdata/user_test.json"

// The usernames and the passwords here correspond
// to the username added from the testdata json file
// to the User Collection
const testAdminUsername string = "admin"
const testAdminUserPassword string = "changeMe"
const testStaffUsername string = "staff"
const testStaffUserPassword string = "tellTheTruth"
const testCustomerUsername string = "customer1"
const testCustomerUserPassword string = "password"

// setup Setup the database for testing by creating a connection to the
// database and returning a handle to the UserService. If desired
// via the clear flag the current user collection entires can be
// dropped. Setting the load flag causes the predefined user collection
// entires in TestUserFilename to be inserted into the user collection.
func SetupUser(t *testing.T, clear bool, load bool) *db.UserService {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	assert.NoError(err)
	
	config := db.Config{Database: testDatabase, URL}
	client := mongo.
	us, err := db.NewUserService(context.TODO(), &config)
	assert.NoError(t, err)
	if clear {
		err = us.DeleteAll()
		assert.NoError(t, err)
	}
	if load {
		err = us.LoadFromFile(testUserFilename)
		assert.NoError(t, err, "Load of json data from %s failed", testUserFilename)
	}
	return us
}

// teardown - perform database teardown to ensure each
// that the database is clean
func TeardownUser(t *testing.T, us *db.UserService) {
	err := us.DeleteAll()
	assert.NoError(t, err)
}

// TestNewUserService perform a simple instantiation test
func TestNewUserService(t *testing.T) {
	// Ensure that we got a handle to the service
	config := db.Config{Database: testDatabase}
	ref, err := db.NewUserService(context.TODO(), &config)
	assert.NoError(t, err)
	assert.NotNil(t, ref)
}

// TestCreateUser test the creation of users.
func TestCreateUser(t *testing.T) {
	userService := SetupUser(t, true, false)
	defer TeardownUser(t, userService)

	// Add a customer
	user := client.User{
		Username: "customer1",
		Password: "password",
	}
	err := userService.Create(&user)
	assert.NoError(t, err)

	// Add a new staff user
	user = client.User{
		Username:  "wwomen",
		Password:  "password",
		Privilege: perm.Staff.String(),
	}
	err = userService.Create(&user)
	assert.NoError(t, err)

	// Attempt to add a different user
	user = client.User{
		Username:  "altAdmin",
		Password:  "changeMe",
		Privilege: perm.Admin.String(),
	}
	err = userService.Create(&user)
	assert.NoError(t, err)
}

// TestCreateDuplicatUser ensure an attempt to add a duplicate user fails
func TestCreateDuplicateUser(t *testing.T) {
	userService := SetupUser(t, true, false)
	defer TeardownUser(t, userService)

	// Add a customer
	user := client.User{
		Username: "customer1",
		Password: "password",
	}
	err := userService.Create(&user)
	assert.NoError(t, err)

	// Attempt to add same user
	err = userService.Create(&user)
	assert.Contains(t, err.Error(), "already exists")
}

// TestValidate ensure that a Validate correctly
// handles a login with correct and incorrect passwords
func TestValidate(t *testing.T) {
	SetupUser(t, true, true)
	userService := SetupUser(t, true, true)

	defer TeardownUser(t, userService)

	// Attempt to login in with the proper credentials
	// NOTE: This test will fail if the credentials
	// in TestUserFilename are changed
	credentials := client.Credentials{
		Username: "customer1",
		Password: "password",
	}
	user, err := userService.Validate(&credentials)
	assert.NoError(t, err)
	assert.True(t, len(user.ID) != 0)
	assert.Equal(t, credentials.Username, user.Username)
	assert.Equal(t, "-", user.Password)
	assert.Equal(t, perm.Basic.String(), user.Privilege)

	// Atempt to login with the improper credentials
	credentials.Password = "wrongPassword"
	user, err = userService.Validate(&credentials)
	assert.Error(t, err)
}

func TestGetUserByName(t *testing.T) {
	userService := SetupUser(t, true, true)
	defer TeardownUser(t, userService)

	// Attempt to retrieve the user populated to the user collection
	retUser, err := userService.GetByUsername(testAdminUsername)
	assert.Equal(t, testAdminUsername, retUser.Username)
	assert.Equal(t, perm.Admin.String(), retUser.Privilege)
	assert.NoError(t, err)

	retUser, err = userService.GetByUsername(testStaffUsername)
	assert.Equal(t, testStaffUsername, retUser.Username)
	assert.Equal(t, perm.Staff.String(), retUser.Privilege)
	assert.NoError(t, err)

	retUser, err = userService.GetByUsername(testCustomerUsername)
	assert.Equal(t, testCustomerUsername, retUser.Username)
	assert.Equal(t, perm.Basic.String(), retUser.Privilege)
	assert.NoError(t, err)

	// Attempt to retrieve a non existent user
	retUser, err = userService.GetByUsername("noexist")
	assert.Error(t, err)
}

func TestGetAllUsers(t *testing.T) {
	userService := SetupUser(t, true, true)
	defer TeardownUser(t, userService)

	// Attempt to retrieve the populated users
	users, err := userService.GetAll()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(users))
}

func TestDeleteUser(t *testing.T) {
	userService := SetupUser(t, true, true)
	defer TeardownUser(t, userService)

	// Attempt to delete an existing user from the database
	user, err := userService.GetByUsername(testCustomerUsername)
	assert.Equal(t, testCustomerUsername, user.Username)
	assert.NoError(t, err)

	err = userService.DeleteUserData(user.ID)
	assert.NoError(t, err)

	// Attempt to delete a non existent user
	err = userService.DeleteUserData(primitive.NewObjectID().Hex())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete")

	// Attempt delete with bad ID
	err = userService.DeleteUserData("4dddkalkdlajbeee")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}
