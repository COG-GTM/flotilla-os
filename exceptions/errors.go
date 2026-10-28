package exceptions

//
// MalformedInput describes malformed or otherwise incorrect input
//
type MalformedInput struct {
	ErrorString string
}

func (e MalformedInput) Error() string {
	return e.ErrorString
}

//
// ConflictingResource describes a conflict case:
// eg. definition already exists, reserved fields
//
type ConflictingResource struct {
	ErrorString string
}

func (e ConflictingResource) Error() string {
	return e.ErrorString
}

//
// ResourceMissing describes case where a resource does not exist
// eg. missing definition or run or no image found
//
type MissingResource struct {
	ErrorString string
}

func (e MissingResource) Error() string {
	return e.ErrorString
}

//
// Unauthenticated describes a request without valid credentials
//
type Unauthenticated struct {
	ErrorString string
}

func (e Unauthenticated) Error() string {
	return e.ErrorString
}

//
// Forbidden describes an authenticated caller that lacks the required role
//
type Forbidden struct {
	ErrorString string
}

func (e Forbidden) Error() string {
	return e.ErrorString
}
