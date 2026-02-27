package queue

type TaskNotFoundError struct {
	Id string
}

func (e *TaskNotFoundError) Error() string {
	return "Task not found: " + e.Id
}

type GenericError struct {
	Inner error
}

func (e *GenericError) Error() string {
	return "An unexpected error occurred: " + e.Inner.Error()
}

func (e *GenericError) Unwrap() error {
	return e.Inner
}
