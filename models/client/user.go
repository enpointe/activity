package client

// UserUpdate the model used to update a user
type UserUpdate struct {
	ID        string `json:"id,unique" example:"5db8e02b0e7aa732afd7fbc4"`
	Username  string `json:"username,unique" example:"admin"`
	Password  string `json:"password,omitempty" example:"myPassword"`
	Privilege string `json:"privilege,omitempty" example:"admin"`
}

// UserCreate model used to create a user
type UserCreate struct {
	Username  string `json:"username,unique" example:"admin"`
	Password  string `json:"password,omitempty" example:"myPassword"`
	Privilege string `json:"privilege,omitempty" example:"admin"`
}

// UserInfo model used to return information about a given user
type UserInfo struct {
	ID        string `json:"id,unique" example:"5db8e02b0e7aa732afd7fbc4"`
	Username  string `json:"username,unique" example:"admin"`
	Privilege string `json:"privilege,omitempty" example:"admin"`
}
