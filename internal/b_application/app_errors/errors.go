package appErrors

type Kind int

type AppError struct {
	kind Kind
	msg  string
}

const (
	KindValidation Kind = iota
	KindNotFound
	KindAlreadyExists
	KindInternalServer
)

func (e *AppError) Error() string {
	return e.msg
}

func (e *AppError) Kind() Kind {
	return e.kind
}

func Validation(msg string) *AppError {
	return &AppError{kind: KindValidation, msg: msg}
}

func NotFound(msg string) *AppError {
	return &AppError{kind: KindNotFound, msg: msg}
}

func AlreadyExists(msg string) *AppError {
	return &AppError{kind: KindAlreadyExists, msg: msg}
}

func InternalServer(msg string) *AppError {
	return &AppError{kind: KindInternalServer, msg: msg}
}
