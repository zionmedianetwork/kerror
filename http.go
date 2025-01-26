package kerror

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type httpErrorResponse struct {
	Code      int               `json:"-"`
	Kind      string            `json:"kind"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id"`
	Fields    map[string]string `json:"fields,omitempty"`
}

func HTTPErrorResponse(c echo.Context, e error) error {
	var res httpErrorResponse
	switch k := e.(type) {
	case *echo.HTTPError:
		res.Code = k.Code
		res.Kind = ErrEchoServer.String()
		res.Message = fmt.Sprintf("%v", k.Message)
	default:
		// try to cast that thing to KissaError
		// if not we'll create an InternalServerError
		// from it
		ke, ok := e.(KisaError)
		if !ok {
			ke = InternalServerError(e)
		}

		// Get the concrete object
		kerror, ok := ke.(*Kerror)
		if !ok {
			log.Panic("wrong error")
		}

		res.Code = kerror.Code()
		res.Kind = kerror.Kind
		res.Message = kerror.GetMessage()

		// If this is a validation error
		// pass the add the fields to the
		// response error payload
		if Is(ke, ErrValidation) {
			res.Fields = kerror.GetValidationFields()
		}

		// Overwrite error message in case of
		// internal_error
		if Is(ke, ErrInternalServer) {
			res.Message = "Internal server error!"
		}

	}

	res.RequestID = c.Response().Writer.Header().Get("X-Request-Id")

	return c.JSON(res.Code, map[string]httpErrorResponse{"error": res})
}

func BindError(err error, message string) KisaError {
	if kerror, ok := err.(*echo.HTTPError); ok {
		return InvalidRequestError(fmt.Sprintf("error byding type %s, %v", message, kerror.Message), nil)
	}

	return InvalidRequestError(fmt.Sprintf("error binding type %s, %v", message, err.Error()), nil)
}

func UnauthorizedError(message string, err error) KisaError {
	e := newKisa(
		message,
		ErrUnauthorized.String(),
		http.StatusUnauthorized,
		err,
	)

	return e
}

func ForbidenError(message string, err error) KisaError {
	e := newKisa(
		message,
		ErrForbidden.String(),
		http.StatusForbidden,
		err,
	)

	return e
}

func NoTokenError(err error) KisaError {
	e := newKisa(
		"No token found in request header.",
		ErrNoToken.String(),
		http.StatusUnauthorized,
		err,
	)

	return e
}

func InvalidTokenError(err error) KisaError {
	e := newKisa(
		"The provided token is not valid.",
		ErrInvalidToken.String(),
		http.StatusUnauthorized,
		err,
	)

	return e
}

func InvalidJWTClaimsError(err error) KisaError {
	e := newKisa(
		"The JWT claims are not valid.",
		ErrInvalidJWTClaims.String(),
		http.StatusUnauthorized,
		err,
	)

	return e
}

func InvalidApiKeyError(err error) KisaError {
	e := newKisa(
		"The provided api key is not valid.",
		ErrInvalidToken.String(),
		http.StatusUnauthorized,
		err,
	)

	return e
}

func InvalidKeyAttributesError(err error) KisaError {
	e := newKisa(
		"The API Key attributes are not valid.",
		ErrInvalidJWTClaims.String(),
		http.StatusUnauthorized,
		err,
	)

	return e
}

func InvalidAuthMethodError(err error) KisaError {
	e := newKisa(
		"The authentication method is not supported.",
		ErrInvalidJWTClaims.String(),
		http.StatusUnauthorized,
		err,
	)

	return e
}
