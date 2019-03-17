package utils

type Error struct {
	LogErrorMessage    string
	ClientErrorMessage string
}

func (err Error) NewError(_error error, _clientErrorMessage string) *Error {
	return &Error{
		LogErrorMessage:    _error.Error(),
		ClientErrorMessage: _clientErrorMessage,
	}
}

func (err *Error) Error() string {
	return err.LogErrorMessage
}
