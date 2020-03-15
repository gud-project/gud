package gud

// Error is a basic error type for errors that are unique to gud.
type Error struct {
	s string
}

func (e Error) Error() string {
	return e.s
}

var ErrUnstagedChanges = Error{"the index must be empty when checking out"}
var ErrUnsavedChanges = Error{"unsaved changes must be cleaned before checking out"}
