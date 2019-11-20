package controllers_test

import (
	"testing"

	"github.com/enpointe/activity/controllers"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestNewServerService(t *testing.T) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	opt := controllers.DBOptions(clientOptions)

	// Test for successful startup skipping admin user test
	server, err := controllers.NewServerService(true, opt, controllers.DBName(testDatabase))
	assert.NoError(t, err)
	assert.NotNil(t, server)

	// Test for startup failure due to missing admin user
	server, err = controllers.NewServerService(false, opt, controllers.DBName(testDatabase))
	assert.Error(t, err)
	assert.Nil(t, server)
}

// TestAdminCreation When instantiating ServerService ensure the admin
// user is created if requested.
func TestAdminOption(t *testing.T) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	sOptions := []controllers.ServerOption{controllers.DBOptions(clientOptions)}
	sOptions = append(sOptions, controllers.CreateAdminUser([]byte("changeMe")))
	activityServer, err := controllers.NewServerService(false, sOptions...)
	assert.NoError(t, err)
	if activityServer != nil {
		activityServer.DeleteAll()
	}
}
