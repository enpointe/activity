package db

import (
	"fmt"
	"regexp"

	"github.com/enpointe/activity/models/client"
	"github.com/enpointe/activity/perm"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// usernameCheck regular expression pattern for allowed username
//
// NOTE: As this is a test project our username requirements are currently
// dictated in a small part via the faker.UserName() routine and the possible
// combinations that it can return. Currently faker.Username() can return a
// 2 letter username
var usernameCheck = regexp.MustCompile(`^[a-zA-Z0-9._-]{2,30}$`).MatchString

// UsernameMinLength the minium length allowed for a username
const UsernameMinLength int = 2

// UsernameMaxLength the maximum length allowed for a username
const UsernameMaxLength int = 30

// PasswordMinLength the minimum length allowed for a password
const PasswordMinLength int = 6

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
func NewUser(u *client.UserCreate) (*User, error) {
	if !usernameCheck(u.Username) {
		err := fmt.Errorf("invalid username specified, '%s'", u.Username)
		return nil, err
	}
	// NOTE: In a real production environment a stricter password
	// check would need to be done to ensure the user creates a secure
	// password.
	if len(u.Password) < PasswordMinLength {
		err := fmt.Errorf("invalid password specified, minium length is %d", PasswordMinLength)
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

// Convert transform into a client facing UserInfo object
func (u *User) Convert() client.UserInfo {
	return client.UserInfo{
		ID:        u.ID.Hex(),
		Username:  u.Username,
		Privilege: u.Privilege.String(),
	}
}
