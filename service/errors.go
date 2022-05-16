package service

import (
	"errors"
	"fmt"
)

var (
	ErrorNotFound = errors.New("not found")
)

type ValidationError struct {
	ObjectKind string
	Message    string
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("failed to validate %s: %s", ve.ObjectKind, ve.Message)
}
