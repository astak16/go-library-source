package multierror

func Append(err error, errs ...error) *Error {
	switch newErr := err.(type) {
	case *Error:
		if newErr == nil {
			newErr = new(Error)
		}
		for _, e := range errs {
			switch e := e.(type) {
			case *Error:
				if e != nil {
					newErr.Errors = append(newErr.Errors, e.Errors...)
				}
			default:
				if e != nil {
					newErr.Errors = append(newErr.Errors, e)
				}
			}
		}
		return newErr
	default:
		newErrs := make([]error, 0, len(errs)+1)
		if err != nil {
			newErrs = append(newErrs, err)
		}
		newErrs = append(newErrs, errs...)
		return Append(&Error{}, newErrs...)
	}

}
