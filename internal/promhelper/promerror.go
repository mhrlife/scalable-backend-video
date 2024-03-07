package promhelper

type PromError struct {
	status Status
	error  error
}

func (p PromError) Error() string {
	return p.error.Error()
}

func NewPromError(status Status, err error) PromError {
	return PromError{
		status: status,
		error:  err,
	}
}
