package service

import "fmt"

type ValidationError struct {
	ObjectKind string
	Message    string
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("failed to validate %s: %s", ve.ObjectKind, ve.Message)
}
