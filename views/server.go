package views

import (
	"context"
	"log"
	"time"

	"github.com/enpointe/activity/models/db"
)

// NewServer some comment
type NewServer struct {
	Config         db.Config
	ListenAddress  string
	skipAdminCheck bool
}

// Option options for the server that can be passed in by the callee
type Option func(*NewServer)

// DBConfig specify the database configuration to use
func DBConfig(config db.Config) Option {
	return func(s *NewServer) {
		s.Config = config
	}
}

// ListenAddress set the TCP network address for the server to listen on
func ListenAddress(addr string) Option {
	return func(s *NewServer) {
		s.ListenAddress = addr
	}
}

// SkipAdminUserCheck flag to indicate whether to skip the administrator
// user check at startup
func SkipAdminUserCheck(skip bool) Option {
	return func(s *NewServer) {
		s.skipAdminCheck = skip
	}
}

// NewServerService create a server service that can be used to
// instantiate a http server. This option will fail is no
// admin privilege user has been configured
func NewServerService(opts ...Option) *NewServer {
	server := &NewServer{}
	for _, opt := range opts {
		opt(server)
	}

	// In order to configure and use this product at least one admin
	// level user is required. Check to ensure that the admin
	// privilege user exists. If no admin privileged user exists
	// then abort startup
	context, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
	defer cancel()
	userService, err := db.NewUserService(context, &server.Config)
	if err != nil {
		log.Panicf("server startup error, failure connecting to database: %s", err)
	}

	configured, err := userService.AdminUserExists()
	if err != nil {
		log.Panicf("server startup error, unable to confirm whether an admin level user exists: %s", err)
	}
	if !configured {
		log.Panic("server startup error, no 'admin' configured. Please configure 'admin' user via cli interface.")
	}
	return server
}
