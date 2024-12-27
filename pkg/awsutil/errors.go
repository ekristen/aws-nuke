package awsutil

const ErrCodeInvalidAction = "InvalidAction"
const ErrCodeOperationNotPermitted = "OperationNotPermitted"

type ErrSkipRequest string

func (err ErrSkipRequest) Error() string {
	return string(err)
}

type ErrUnknownEndpoint string

func (err ErrUnknownEndpoint) Error() string {
	return string(err)
}
