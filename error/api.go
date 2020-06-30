package error

type ApiError struct {
	Msg string
}

func (ae ApiError) Error() string {
	return ae.Msg
}
