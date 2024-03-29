package client

// Credentials login credentials
type Credentials struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"changeMe"`
}

// PasswordUpdate used to change the password for a given user
type PasswordUpdate struct {
	ID              string `json:"id,unique"`
	NewPassword     string `json:"newPassword"`
	CurrentPassword string `json:"currentPassword"`
}
