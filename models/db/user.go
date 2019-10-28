package db

import (
	"github.com/enpointe/activity/models/client"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// User represents data about the person exercising
type User struct {
	ID       primitive.ObjectID `bson:"_id,unique"`
	UserName string             `bson:"user_id,unique"`
	Password string             `bson:"password"`
}

// NewUser transforms the web facing User structure
// to a database compatible User structure. The ID field is
// automaticly set to a primitive.NewObjectID() any passed
// in value is ignored
func NewUser(u *client.User) (*User, error) {
	user := User{
		ID:       primitive.NewObjectID(),
		UserName: u.Username,
	}
	err := user.setHashedPassword(u.Password)
	return &user, err
}

func (u *User) setHashedPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	u.Password = string(bytes)
	return nil
}

func (u *User) comparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
