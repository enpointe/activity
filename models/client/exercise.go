package client

// Exercise the model exposed via the web interface
type Exercise struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ExerciseService functions available to Exercise
type ExerciseService interface {
	CreateExercise(e *Exercise) error
	DeleteExercise(id string) error
	UpdateExercise(e *Exercise) (*Exercise, error)
	GetExercise(id string) (*Exercise, error)
	GetAllExercises() (*[]Exercise, error)
}
