package gud

type Error struct {
	s string
}

func (e Error) Error() string {
	return e.s
}
