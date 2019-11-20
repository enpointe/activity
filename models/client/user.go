package client

// User the user model for json http interface
// The ID field is the identifier field for the
// record associated with the User structure.
type User struct {
	ID        string `json:"id,unique"`
	Username  string `json:"username,unique"`
	Password  string `json:"password,omitempty"`
	Privilege string `json:"privilege,omitempty"`
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
