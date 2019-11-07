package views

import (
	"context"
	"fmt"
	"time"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/models/db"
	"github.com/enpointe/activity/perm"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DefaultDatabase the database name to use as the default
const DefaultDatabase string = "activity"

// ServerService some comment
type ServerService struct {
	dbName      string
	adminPasswd []byte
	dbClOpts    *options.ClientOptions
	client      *mongo.Client
	Database    *mongo.Database
}

// ServerOption options for the server that can be passed in by the callee
type ServerOption func(*ServerService)

// DBOptions specify the database client options
func DBOptions(options *options.ClientOptions) ServerOption {
	return func(s *ServerService) {
		s.dbClOpts = options
	}
}

// DBName specifies the database name to use
func DBName(name string) ServerOption {
	return func(s *ServerService) {
		s.dbName = name
	}
}

// CreateAdminUser create the user "admin" and assign it the specified password.
// If the admin user already exists the password will be updated to the
// specified password
func CreateAdminUser(passwd []byte) ServerOption {
	return func(s *ServerService) {
		s.adminPasswd = passwd
	}
}

// NewServerService create a server service that can be used to
// instantiate a http server. This option will fail is no
// admin privilege user has been configured
func NewServerService(skipAdminCheck bool, opts ...ServerOption) (*ServerService, error) {
	log.Debug("Creating ServerService")
	server := &ServerService{dbName: DefaultDatabase}
	for _, opt := range opts {
		opt(server)
	}
	if server.dbClOpts == nil {
		server.dbClOpts = options.Client().ApplyURI("mongodb://localhost:27017")
	}

	mClient, err := mongo.NewClient(server.dbClOpts)
	if err != nil {
		err = fmt.Errorf("failed to open a connection to MongoDB, %s", err)
		return nil, err
	}
	server.client = mClient

	ctx, cancel := context.WithTimeout(context.TODO(), 90*time.Second)
	defer cancel()
	err = mClient.Connect(ctx)
	if err != nil {
		return nil, err
	}
	server.Database = mClient.Database(server.dbName)

	if skipAdminCheck {
		return server, nil
	}

	// In order to configure and use this product at least one admin
	// level user is required. Check to ensure that the admin
	// privilege user exists. If no admin privileged user exists
	// then abort startup
	userService, err := db.NewUserService(server.client, server.Database)
	if err != nil {
		err = fmt.Errorf("server startup error, failure connecting to database: %s", err)
		return nil, err
	}
	configured, err := userService.AdminUserExists(ctx)
	if len(server.adminPasswd) > 0 && (!configured || err != nil) {
		// The user requested that we create an admin user.
		// This option is only allowed if an admin user doesn't already exist
		if configured {
			err = fmt.Errorf("admin privilege user already configured, request to create admin user rejected")
			return nil, err
		}
		user := client.User{
			Username:  "admin",
			Password:  string(server.adminPasswd),
			Privilege: perm.Admin.String(),
		}
		id, err := userService.Create(ctx, &user)
		if err != nil {
			err = fmt.Errorf("failed to create admin privilege user: %s", err)
			return nil, err
		}
		log.Infof("Successfully created admin user id %s", id)
		configured = true
	}
	if err != nil {
		err = fmt.Errorf("server startup error, admin privilege check: %s", err)
		return nil, err
	}
	if !configured {
		err = fmt.Errorf("server startup error, no 'admin' privilege user configured")
		return nil, err
	}
	return server, nil
}

// DeleteAll delete all collections in the database
func (s *ServerService) DeleteAll() error {
	return s.Database.Drop(context.TODO())
}

// Shutdown performs any clean up activites related to the running the service.
func (s *ServerService) Shutdown() error {
	return s.client.Disconnect(context.Background())
}
