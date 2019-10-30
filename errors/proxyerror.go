package errors

const (
	ServiceNotFound  = 400
	AuthFailed       = 403
	EncryptionFailed = 500
	DecryptionFailed = 500
	BadRequest       = 400
)

type ProxyError struct {
	Code    int
	Message string
}

func (err ProxyError) Error() string {
	return err.Message
}
