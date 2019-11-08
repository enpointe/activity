package views_test

import (
	"testing"

	"github.com/enpointe/activity/views"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestNewServerService(t *testing.T) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	opt := views.DBOptions(clientOptions)

	// Test for successful startup skipping admin user test
	server, err := views.NewServerService(true, opt, views.DBName(testDatabase))
	assert.NoError(t, err)
	assert.NotNil(t, server)

	// Test for startup failure due to missing admin user
	server, err = views.NewServerService(false, opt, views.DBName(testDatabase))
	assert.Error(t, err)
	assert.Nil(t, server)
}

// TestAdminCreation When instantiating ServerService ensure the admin
// user is created if requested.
func TestAdminOption(t *testing.T) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	sOptions := []views.ServerOption{views.DBOptions(clientOptions)}
	sOptions = append(sOptions, views.CreateAdminUser([]byte("changeMe")))
	activityServer, err := views.NewServerService(false, sOptions...)
	assert.NoError(t, err)
	if activityServer != nil {
		activityServer.DeleteAll()
	}
}
