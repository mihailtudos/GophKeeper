package errors

import "fmt"

func WrapStandardError(op, message string, err error) error {
	return fmt.Errorf("%s - %s: %w", op, message, err)
}
