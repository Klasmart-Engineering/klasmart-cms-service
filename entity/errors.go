package entity

type ErrInvalidArgs error

func IsErrInvalidArgs(err error) bool {
	_, ok := err.(ErrInvalidArgs)
	return ok
}
