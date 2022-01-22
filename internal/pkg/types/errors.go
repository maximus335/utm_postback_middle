package types

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

type ValidationParamsError struct {
	Message string
}

func (e *ValidationParamsError) Error() string {
	return e.Message
}
