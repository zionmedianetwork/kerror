package kerror

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/lib/pq"
	"github.com/zionmedianetwork/logam"
)

type Kind uint8

const (
	ErrUnkown Kind = iota
	ErrInvalidRequest
	ErrNotFound
	ErrResourceConflict
	ErrInternalServer
	ErrValidation
	ErrDatabase
	ErrUnauthorized
	ErrForbidden
	ErrRequestTimeout
	ErrEchoServer
	ErrNoToken
	ErrInvalidToken
	ErrInvalidJWTClaims
)

func (k Kind) String() string {
	switch k {
	case ErrUnkown:
		return "unknown_error"
	case ErrInvalidRequest:
		return "invalid_request_error"
	case ErrNotFound:
		return "item_not_found_error"
	case ErrResourceConflict:
		return "resource_conflict_error"
	case ErrInternalServer:
		return "internal_server_error"
	case ErrValidation:
		return "input_validation_error"
	case ErrDatabase:
		return "database_error"
	case ErrUnauthorized:
		return "authorization_error"
	case ErrForbidden:
		return "forbiden_error"
	case ErrRequestTimeout:
		return "request_timeout_err"
	case ErrEchoServer:
		return "server_error"
	case ErrNoToken:
		return "no_token_error"
	case ErrInvalidToken:
		return "invalid_token_error"
	case ErrInvalidJWTClaims:
		return "invalid_jwt_claims_error"
	default:
		return "unkown_error"
	}
}

type LogContext map[string]interface{}
type validationFields []FieldValidationError
type KisaError interface {
	Error() string
	GetMessage() string
	Code() int
	With(...interface{}) KisaError
	GetValidationFields() map[string]string
	GetKind() string
	setValidationFields(validationFields)
	setLogger(l logam.Logger)
}

type Kerror struct {
	Status  int
	Kind    string
	Message string
	Fields  map[string]string
	context LogContext
	err     error
	logger  logam.Logger
}

func newKisa(m string, k string, s int, err error) KisaError {
	return &Kerror{
		Status:  s,
		Kind:    k,
		Message: m,
		err:     err,
	}
}

func (k Kerror) Error() string {
	return fmt.Sprintf("status: %d - message: %s - errors: %v", k.Status, k.Message, k.err)
}

func (k Kerror) Code() int {
	return k.Status
}

func (k Kerror) GetKind() string {
	return k.Kind
}

func (k Kerror) GetMessage() string {
	return k.Message
}

func (k Kerror) GetValidationFields() map[string]string {
	return k.Fields
}

func (k Kerror) Unwrap() error {
	return k.err
}

func (k *Kerror) With(contextArgs ...interface{}) KisaError {
	if contextArgs == nil {
		return k
	}
	numArgs := len(contextArgs)
	if numArgs == 1 {
		k.Message = fmt.Sprintf("%s, %v", k.Message, contextArgs[0])
	}

	k.context = make(LogContext)
	for i := 0; i < numArgs; i += 2 {
		contextKey := contextArgs[i].(string)
		contextVal := contextArgs[i+1]
		k.context[contextKey] = contextVal
	}

	k.messageFromContext()

	if k.logger != nil {
		defer k.logger.Errorw(k.Message, contextArgs...)
	}

	return k
}

func (k *Kerror) messageFromContext() {
	if k.context != nil {
		for a, v := range k.context {
			if a == "what" {
				k.Message = k.Message + ", " + v.(string)
				break
			}
		}
	}
}

func (k *Kerror) setValidationFields(vf validationFields) {
	fields := map[string]string{}

	for _, f := range vf {
		fields[f.FieldName] = f.Error.Error()
	}

	k.Fields = fields
}

func (k *Kerror) setLogger(l logam.Logger) {
	k.logger = l
}

func InvalidRequestError(message string, err error) KisaError {
	e := newKisa(
		message,
		ErrInvalidRequest.String(),
		http.StatusBadRequest,
		err)

	return e
}

func Is(e KisaError, k Kind) bool {
	err, ok := e.(*Kerror)
	if !ok {
		return false
	}

	return err.Kind == k.String()
}

func NotFoundError(message string, err error) KisaError {
	e := newKisa(
		message,
		ErrNotFound.String(),
		http.StatusNotFound,
		err,
	)

	return e
}

func ResourceConflictError(message string, err error) KisaError {
	e := newKisa(
		message,
		ErrResourceConflict.String(),
		http.StatusConflict,
		err,
	)

	return e
}
func InternalServerError(err error) KisaError {
	return newKisa(
		"this is an internal server error! our team has been notified",
		ErrInternalServer.String(),
		http.StatusInternalServerError,
		err,
	)
}

func DatabaseError(err error, log logam.Logger) KisaError {
	if errors.Is(err, sql.ErrNoRows) {
		e := NotFoundError("the resource could not be found in the database", err)

		return e
	}

	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		// Duplicate entry
		case "23505":
			e := ResourceConflictError("this entry already existed in the database", err)
			e.setLogger(log)

			return e
		default:
			e := InternalServerError(nil)
			e.setLogger(log)

			return e
		}
	}

	return InternalServerError(err)

}

func ValidationError(err validationFields) KisaError {
	e := newKisa(
		"could not validate one or more of the submitted fields",
		ErrValidation.String(),
		http.StatusBadRequest,
		nil,
	)

	e.setValidationFields(err)

	return e
}

// ToKerror takes a generic KisaError
// and return a structured Kerrok
// instance
func Tokerror(e error) *Kerror {
	err := e.(KisaError)
	if err != nil {
		kerr := &Kerror{
			Status:  err.Code(),
			Kind:    err.GetKind(),
			Message: err.GetMessage(),
		}

		if Is(err, ErrValidation) {
			kerr.Fields = err.GetValidationFields()
		}

		return kerr
	}

	return nil
}
