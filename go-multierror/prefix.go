package multierror

import (
	"fmt"
)

func Prefix(err error, prefix string) error {
	if err == nil {
		return nil
	}

	switch err := err.(type) {
	case *Error:
		if err == nil {
			err = new(Error)
		}

		for i, e := range err.Errors {
			err.Errors[i] = fmt.Errorf("%s %s", prefix, e)
		}

		return err
	default:
		return fmt.Errorf("%s %s", prefix, err)
	}
}
