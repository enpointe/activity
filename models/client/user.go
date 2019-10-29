package client

// User the user model for json http interface
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Privilege string `json:"privilege"`
}

// UserService functions to associate with User struct
type UserService interface {
	CreateUser(u *User) error
	DeleteAllUserData(u *User) error
	GetUserByUsername(username string) (error, User)
	Validate(c Credentials) (error, User)
	GetAllUsers() ([]*User, error)
}
