package db_test

import (
	"context"
	"testing"

	"github.com/enpointe/activity/perm"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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
const testAdminID string = "5db8e02b0e7aa732afd7fbc4"
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
	clientOptions := options.Client().ApplyURI(testDatabaseURL)
	ctx := context.TODO()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	assert.NoError(t, err)

	database := client.Database(testDatabase)
	us, err := db.NewUserService(database)
	assert.NoError(t, err)
	if clear {
		err = us.DeleteAll(ctx)
		assert.NoError(t, err)
	}
	if load {
		err = us.LoadFromFile(ctx, testUserFilename)
		assert.NoError(t, err, "Load of json data from %s failed", testUserFilename)
	}
	return us
}

// teardown - perform database teardown to ensure each
// that the database is clean
func TeardownUser(t *testing.T, us *db.UserService) {
	err := us.DeleteAll(context.TODO())
	assert.NoError(t, err)
}

// TestCreateUser test the creation of users.
func TestCreateUser(t *testing.T) {
	userService := SetupUser(t, true, false)
	defer TeardownUser(t, userService)
	ctx := context.TODO()

	// Add a customer
	user := client.User{
		Username: "customer1",
		Password: "password",
	}
	id, err := userService.Create(ctx, &user)
	assert.NoError(t, err)
	assert.NotNil(t, id)

	// Add a new staff user
	user = client.User{
		Username:  "wwomen",
		Password:  "password",
		Privilege: perm.Staff.String(),
	}
	id, err = userService.Create(ctx, &user)
	assert.NoError(t, err)
	assert.NotNil(t, id)

	// Attempt to add a different user
	user = client.User{
		Username:  "altAdmin",
		Password:  "changeMe",
		Privilege: perm.Admin.String(),
	}
	id, err = userService.Create(ctx, &user)
	assert.NoError(t, err)
	assert.NotNil(t, id)
}

func TestCreateUserFailures(t *testing.T) {
	userService := SetupUser(t, false, false)
	defer TeardownUser(t, userService)
	ctx := context.TODO()

	// Test no user specified
	user := client.User{
		Password: "password",
	}
	id, err := userService.Create(ctx, &user)
	assert.Error(t, err)
	assert.Empty(t, id)

	// Test no password specified
	user = client.User{
		Username: "customer1",
	}
	_, err = userService.Create(ctx, &user)
	assert.Error(t, err)
}

// TestCreateDuplicatUser ensure an attempt to add a duplicate user fails
func TestCreateDuplicateUser(t *testing.T) {
	userService := SetupUser(t, true, false)
	defer TeardownUser(t, userService)
	ctx := context.TODO()

	// Add a customer
	user := client.User{
		Username: "customer1",
		Password: "password",
	}
	id, err := userService.Create(ctx, &user)
	assert.NoError(t, err)
	assert.NotNil(t, id)

	// Attempt to add same user
	id, err = userService.Create(ctx, &user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.NotNil(t, id)
}

// TestValidate ensure that a Validate correctly
// handles a login with correct and incorrect passwords
func TestValidate(t *testing.T) {
	SetupUser(t, true, true)
	userService := SetupUser(t, true, true)
	defer TeardownUser(t, userService)
	ctx := context.TODO()

	// Attempt to login in with the proper credentials
	// NOTE: This test will fail if the credentials
	// in TestUserFilename are changed
	credentials := client.Credentials{
		Username: "customer1",
		Password: "password",
	}
	user, err := userService.Validate(ctx, &credentials)
	assert.NoError(t, err)
	assert.True(t, len(user.ID) != 0)
	assert.Equal(t, credentials.Username, user.Username)
	assert.Equal(t, "-", user.Password)
	assert.Equal(t, perm.Basic.String(), user.Privilege)

	// Atempt to login with the improper credentials
	credentials.Password = "wrongPassword"
	user, err = userService.Validate(ctx, &credentials)
	assert.Error(t, err)
}

func TestGetUserByName(t *testing.T) {
	userService := SetupUser(t, true, true)
	defer TeardownUser(t, userService)
	ctx := context.TODO()

	// Attempt to retrieve the user populated to the user collection
	retUser, err := userService.GetByUsername(ctx, testAdminUsername)
	assert.Equal(t, testAdminUsername, retUser.Username)
	assert.Equal(t, perm.Admin.String(), retUser.Privilege)
	assert.NoError(t, err)

	retUser, err = userService.GetByUsername(ctx, testStaffUsername)
	assert.Equal(t, testStaffUsername, retUser.Username)
	assert.Equal(t, perm.Staff.String(), retUser.Privilege)
	assert.NoError(t, err)

	retUser, err = userService.GetByUsername(ctx, testCustomerUsername)
	assert.Equal(t, testCustomerUsername, retUser.Username)
	assert.Equal(t, perm.Basic.String(), retUser.Privilege)
	assert.NoError(t, err)

	// Attempt to retrieve a non existent user
	retUser, err = userService.GetByUsername(ctx, "noexist")
	assert.Error(t, err)
}

func TestGetByName(t *testing.T) {
	userService := SetupUser(t, true, true)
	defer TeardownUser(t, userService)
	ctx := context.TODO()

	// Attempt to retrieve the user populated to the user collection
	retUser, err := userService.GetByID(ctx, testAdminID)
	assert.Equal(t, testAdminUsername, retUser.Username)
	assert.Equal(t, perm.Admin.String(), retUser.Privilege)
	assert.NoError(t, err)

	// Attempt to retrieve a non existent user
	retUser, err = userService.GetByID(ctx, "noexist")
	assert.Error(t, err)
}

func TestGetAllUsers(t *testing.T) {
	userService := SetupUser(t, true, true)
	defer TeardownUser(t, userService)

	// Attempt to retrieve the populated users
	users, err := userService.GetAll(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 3, len(users))
}

func TestDeleteUser(t *testing.T) {
	userService := SetupUser(t, true, true)
	defer TeardownUser(t, userService)
	ctx := context.TODO()

	// Attempt to delete an existing user from the database
	user, err := userService.GetByUsername(ctx, testCustomerUsername)
	assert.Equal(t, testCustomerUsername, user.Username)
	assert.NoError(t, err)

	cnt, err := userService.DeleteUserData(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, cnt)

	// Attempt to delete a non existent user
	cnt, err = userService.DeleteUserData(ctx, primitive.NewObjectID().Hex())
	assert.NoError(t, err)
	assert.Equal(t, 0, cnt)

	// Attempt delete with bad ID
	cnt, err = userService.DeleteUserData(ctx, "4dddkalkdlajbeee")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
	assert.Equal(t, 0, cnt)
}

func TestUpdateUser(t *testing.T) {
	service := SetupUser(t, true, true)
	ctx := context.TODO()

	testUsername := "admin2"
	user := client.User{
		ID:       testAdminID,
		Username: testUsername,
		Password: testAdminUserPassword,
	}
	err := service.Update(ctx, &user)
	assert.NoError(t, err)
	// Fetch the user and check to make sure the password
	// was updated
	u, err := service.GetByID(ctx, testAdminID)
	assert.NoError(t, err)
	assert.Equal(t, u.Username, testUsername)
}

func TestUpdateFailures(t *testing.T) {
	service := SetupUser(t, false, false)
	ctx := context.TODO()

	user := client.User{
		ID:       "6666e02b0e7aa732afd7fbc4",
		Username: testCustomerUsername,
		Password: testCustomerUserPassword,
	}
	// Test non existent user
	err := service.Update(ctx, &user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update")

	// Test base ID
	user.ID = ""
	err = service.Update(ctx, &user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")

}

func TestLoadUserFromFile(t *testing.T) {
	// Test load from non existant file
	service := SetupUser(t, false, false)
	ctx := context.TODO()
	err := service.LoadFromFile(ctx, "someFile")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")

	// Load in a exercise without a ID specified. Check to make sure
	// an ID is generated
	err = service.LoadFromFile(ctx, "testdata/user_load_test.json")
	assert.NoError(t, err)
	if err != nil {
		ex, err := service.GetByUsername(ctx, "customer1")
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

func TestAdminUserCheckPresent(t *testing.T) {
	service := SetupUser(t, true, true)
	ctx := context.TODO()
	result := service.AdminUserExists(ctx)
	assert.True(t, result)
}

func TestAdminUserCheckNotPresent(t *testing.T) {
	service := SetupUser(t, true, false)
	ctx := context.TODO()
	result := service.AdminUserExists(ctx)
	assert.False(t, result)

}
