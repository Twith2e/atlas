package errors

type TermiiError struct {
	Status  int
	Message string
}

func (e *TermiiError) Error() string {
	return e.Message
}
