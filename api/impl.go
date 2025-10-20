package api

// ensure that we've conformed to the `ServerInterface` with
// a compile-time check
var _ StrictServerInterface = (*Server)(nil)

type Server struct{}

type EmptyInternalServerError struct{}

func (e *EmptyInternalServerError) Error() string {
	return "internal server error"
}
