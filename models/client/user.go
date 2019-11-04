package client

// User the user model for json http interface
// The ID field is the identifier field for the 
// record associated with the User structure.
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Privilege string `json:"privilege"`
}

// UserService functions to associate with User struct
type UserService interface {
	Create(u *User) error
	DeleteUserData(u *User) error
	GetAll() ([]*User, error)
	GetByID(id string) (*User, error)
	GetByUsername(username string) (*User, error)
	Update(u *User) error
	Validate(c *Credentials) (*User, error)
}
