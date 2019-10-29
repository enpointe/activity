package db_test

import (
	"testing"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/enpointe/activity/perm"
	"gotest.tools/assert"
)

func TestNewUser(t *testing.T) {
	// Add a new user
	user := client.User{
		Username:  "wwomen",
		Password:  "tellTheTruth",
		Privilege: "admin",
	}
	dbUser, err := db.NewUser(&user)
	assert.NilError(t, err)
	assert.Assert(t, len(dbUser.ID) > 0)
	assert.Assert(t, dbUser.Password != user.Password)
	assert.Equal(t, dbUser.Privilege, perm.Admin)

	user.Password = ""
	dbUser, err = db.NewUser(&user)
	assert.NilError(t, err)

}
