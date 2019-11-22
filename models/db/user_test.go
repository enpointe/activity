package db_test

import (
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/stretchr/testify/assert"
	"syreclabs.com/go/faker"
)

// TestNewUser test against faker to ensure no username or password failures
func TestNewUser(t *testing.T) {
	for i := 0; i < 100; i++ {
		internet := faker.Internet()
		user := client.UserCreate{
			Username: internet.UserName(),
			Password: internet.Password(db.PasswordMinLength, 30),
		}
		_, err := db.NewUser(&user)
		assert.NoError(t, err)
	}
}

func TestNewUserValidationFailures(t *testing.T) {
	// In a real production application more detail would be performed here
	// to ensure that our username and password regex pattern are catching
	// all the possible invalid combinations that we wish to detect.

	user := client.UserCreate{ // Username too short
		Username:  "a",
		Password:  faker.Internet().Password(db.PasswordMinLength, 30),
		Privilege: "admin",
	}
	dbUser, err := db.NewUser(&user)
	assert.Nil(t, dbUser, "Expected failure for username too short")
	assert.NotNil(t, err)
	if err != nil {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid username")
	}

	user = client.UserCreate{ // Password too short
		Username:  faker.Internet().UserName(),
		Password:  "ab",
		Privilege: "admin",
	}
	dbUser, err = db.NewUser(&user)
	assert.Nil(t, dbUser, "Expected failure for password too short")
	assert.NotNil(t, err)
	if err != nil {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid password")
	}
}
