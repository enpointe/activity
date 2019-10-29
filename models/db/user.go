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
var usernameCheck = regexp.MustCompile(`^[a-z0-9_-]{4,16}$`).MatchString

// User privileges information
type User struct {
	ID        primitive.ObjectID `bson:"_id,unique"`
	Username  string             `bson:"user_id,unique"`
	Password  string             `bson:"password"`
	Privilege perm.Privilege     `bson:"privilege"` // admin, staff, user
}

// NewUser transforms the web facing User structure
// to a database compatible User structure. The ID field is
// automaticly set to a primitive.NewObjectID() any passed
// in value is ignored.
func NewUser(u *client.User) (*User, error) {
	if !usernameCheck(u.Username) {
		err := fmt.Errorf("invalid username specified. Username must be between 4-16 characters and composed of characters: a-z, A-Z, 0-9, -, and _")
		return nil, err
	}
	// Only enforce that password is at least 8 characters long and contains no spaces
	u.Password = strings.TrimSpace(u.Password)
	if len(u.Password) < 8 {
		err := fmt.Errorf("invalid password specified, minimum length is 8 characters")
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
