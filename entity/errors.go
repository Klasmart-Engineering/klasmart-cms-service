package entity

type InvalidArgsError error

func IsInvalidArgsError(err error) bool {
	_, ok := err.(InvalidArgsError)
	return ok
}

type ConflictError error

func IsConflictError(err error) bool {
	_, ok := err.(ConflictError)
	return ok
}
