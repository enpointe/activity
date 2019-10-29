package db_test

import (
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/enpointe/activity/perm"
	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	// Add a new user
	user := client.User{
		Username:  "wwomen",
		Password:  "tellTheTruth",
		Privilege: "admin",
	}
	dbUser, err := db.NewUser(&user)

	assert.Nil(t, err)
	assert.Greater(t, len(dbUser.ID), 0)
	assert.NotEqual(t, dbUser.Password, user.Password)
	assert.Equal(t, user.Privilege, perm.Admin.String())

	user.Password = ""
	dbUser, err = db.NewUser(&user)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "minimum length")

}
