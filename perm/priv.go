package perm

import "fmt"

// Privilege defines privilege level
type Privilege int8

const (
	// Basic lowest level privilege
	Basic Privilege = 0
	// Staff only allow staff level operations
	Staff Privilege = 1
	// Admin all operations allowed
	Admin Privilege = 2
)

// String convert privilge to a string
func (p Privilege) String() string {
	names := [...]string{
		"basic",
		"staff",
		"admin",
	}
	if p < Basic || p > Admin {
		return "basic"
	}
	fmt.Printf("P=%d\n", p)
	return names[p]
}

// Grants does the current privilege meet or exceed the required privilege
func (p Privilege) Grants(required Privilege) bool {
	if required < Basic || p > Admin {
		return false
	}
	return required <= p
}

// Convert convert privilege string to const value.
// If value is unrecoginized the lowest level privilege (Basic)
// will be returned
func Convert(privilege string) Privilege {
	switch privilege {
	case "admin":
		return Admin
	case "staff":
		return Staff
	default:
		return Basic
	}
}
