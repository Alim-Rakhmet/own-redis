package internal

import "errors"

var (
	ErrWrongNumOfArgs = errors.New("(error) Wrong number of arguments")
	ErrUnknownCommand = errors.New("(error) Unkown command")
)
