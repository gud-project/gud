package gud

// Error is a basic error type for errors that are unique to gud.
type Error struct {
	s string
}

func (e Error) Error() string {
	return e.s
}

var AddedUnmodifiedFileError = Error{"the added file has not been modified"}
