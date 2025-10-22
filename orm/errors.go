package orm

type DatabaseError struct {
	Inner error
}

func (e *DatabaseError) Error() string {
	return "Gorm returned an error: " + e.Inner.Error()
}

func (e *DatabaseError) Unwrap() error {
	return e.Inner
}

type NotFoundError struct {
	Search string
}

func (e *NotFoundError) Error() string {
	return "Record not found for search: " + e.Search
}

type ConflictError struct {
	Conflict string
}

func (e *ConflictError) Error() string {
	return "Conflict error for: " + e.Conflict
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
