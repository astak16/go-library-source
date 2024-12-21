package multierror

import "errors"

type Error struct {
	Errors      []error
	ErrorFormat ErrorFormatFunc
}

func (e *Error) Error() string {
	fn := e.ErrorFormat
	if fn == nil {
		fn = ListFormatFunc
	}

	return fn(e.Errors)
}

func (e *Error) ErrorOrNil() error {
	if e == nil {
		return nil
	}
	if len(e.Errors) == 0 {
		return nil
	}

	return e
}

func (e *Error) WrappedErrors() []error {
	if e == nil {
		return nil
	}
	return e.Errors
}

func (e *Error) Unwrap() error {
	// If we have no errors then we do nothing
	if e == nil || len(e.Errors) == 0 {
		return nil
	}

	// If we have exactly one error, we can just return that directly.
	if len(e.Errors) == 1 {
		return e.Errors[0]
	}

	// Shallow copy the slice
	errs := make([]error, len(e.Errors))
	copy(errs, e.Errors)
	return chain(errs)
}

type chain []error

func (e chain) Error() string {
	return e[0].Error()
}

func (e chain) Unwrap() error {
	if len(e) == 1 {
		return nil
	}

	return e[1:]
}

func (e chain) As(target interface{}) bool {
	return errors.As(e[0], target)
}

// // Is implements errors.Is by comparing the current value directly.
func (e chain) Is(target error) bool {
	return errors.Is(e[0], target)
}
