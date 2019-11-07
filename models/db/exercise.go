package db

import (
	"strings"

	"github.com/enpointe/activity/models/client"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exercise represents information for a type of exercise
type Exercise struct {
	ID          primitive.ObjectID `bson:"_id,unique,omitempty" json:"_id,omitempty"`
	Name        string             `bson:"name,omitempty" json:"name,omitempty"`
	Description string             `bson:"description,omitempty" json:"description"`
}

// NewExercise transforms the web facing Exercise structure
// to a database compatible Exercise structure. The ID field is
// automatically set to a primitive.NewObjectID() any passed
// in value is ignored
func NewExercise(e *client.Exercise) *Exercise {
	exercise := Exercise{
		ID:          primitive.NewObjectID(),
		Name:        strings.TrimSpace(e.Name),
		Description: strings.TrimSpace(e.Description),
	}
	return &exercise
}

// Convert transform into a client facing Exercise object
func (e *Exercise) Convert() client.Exercise {
	return client.Exercise{
		ID:          e.ID.Hex(),
		Name:        e.Name,
		Description: e.Description,
	}
}
