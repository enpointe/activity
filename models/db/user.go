package db

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/perm"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// UsernameCheck regular expression pattern for allowed username
var usernameCheck = regexp.MustCompile(`^[A-Za-z0-9_-]{4,16}$`).MatchString

// User privileges information
type User struct {
	ID        primitive.ObjectID `bson:"_id,unique,omitempty" json:"_id,omitempty"`
	Username  string             `bson:"user_id,unique,omitempty" json:"user_id,omitempty"`
	Password  string             `bson:"password,omitempty" json:"password,omitempty"`
	Privilege perm.Privilege     `bson:"privilege,omitempty" json:"privilege,omitempty"` // admin, staff, user
}

// NewUser transforms the web facing User structure
// to a database compatible User structure. The ID field is
// automatically set to a primitive.NewObjectID() any passed
// in ID value is ignored. Username and Password fields
// are checked for correctness.
func NewUser(u *client.User) (*User, error) {
	if !usernameCheck(u.Username) {
		err := fmt.Errorf("invalid username specified. Username must be between 4-16 characters and composed of characters: a-z, A-Z, 0-9, -, and _")
		return nil, err
	}
	user := User{
		ID:        primitive.NewObjectID(),
		Username:  u.Username,
		Privilege: perm.Convert(u.Privilege),
	}

	err := user.setHashedPassword(u.Password)
	return &user, err
}

func (u *User) setHashedPassword(password string) error {
	// Only enforce that password is at least 8 characters long and contains no spaces
	password = strings.TrimSpace(password)
	if len(password) < 8 {
		err := fmt.Errorf("invalid password specified, minimum length is 8 characters")
		return err

	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(bytes)
	return nil
}

func (u *User) comparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// Convert transform into a client facing User object
// Password is converted to "-"
func (u *User) Convert() client.User {
	return client.User{
		ID:        u.ID.Hex(),
		Username:  u.Username,
		Password:  "-",
		Privilege: u.Privilege.String(),
	}
}
