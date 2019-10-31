package client

// Exercise the model exposed via the web interface
type Exercise struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ExerciseService functions available to Exercise
type ExerciseService interface {
	Create(e *Exercise) error
	Delete(id string) error
	GetAll() ([]*Exercise, error)
	GetByID(id string) (*Exercise, error)
	GetByName(name string) (*Exercise, error)
	Update(e *Exercise) error
}
