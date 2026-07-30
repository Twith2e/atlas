package response

import (
	appErr "atlas/internal/errors"
	"net/http"
)

type ErrorMapping struct {
	Status int
	Error  APIError
}

func MapError(err error) ErrorMapping {
	switch err {
	case appErr.ErrInvalidTermiiConfig:
		return ErrorMapping{
			Status: http.StatusInternalServerError,
			Error:  APIError{Code: "INVALID_TERMII_CONFIG", Message: appErr.ErrInvalidTermiiConfig.Error()},
		}
	case appErr.ErrTermiiSendFailed:
		return ErrorMapping{
			Status: http.StatusInternalServerError,
			Error:  APIError{Code: "TERMII_SEND_FAILED", Message: appErr.ErrTermiiSendFailed.Error()},
		}
	case appErr.ErrInvalidPassword:
		return ErrorMapping{
			Status: http.StatusUnprocessableEntity,
			Error:  APIError{Code: "INVALID_PASSWORD", Message: appErr.ErrInvalidPassword.Error()},
		}
	case appErr.ErrInvalidPasswordLength:
		return ErrorMapping{
			Status: http.StatusUnprocessableEntity,
			Error:  APIError{Code: "INVALID_PASSWORD_LENGTH", Message: appErr.ErrInvalidPasswordLength.Error()},
		}
	case appErr.ErrInvalidPhoneNumber:
		return ErrorMapping{
			Status: http.StatusUnprocessableEntity,
			Error:  APIError{Code: "INVALID_PHONE_NUMBER", Message: appErr.ErrInvalidPhoneNumber.Error()},
		}
	case appErr.ErrUserAlreadyExists:
		return ErrorMapping{
			Status: http.StatusConflict,
			Error:  APIError{Code: "USER_ALREADY_EXISTS", Message: appErr.ErrUserAlreadyExists.Error()},
		}
	case appErr.ErrPasswordMismatch:
		return ErrorMapping{
			Status: http.StatusUnprocessableEntity,
			Error:  APIError{Code: "PASSWORD_MISMATCH", Message: appErr.ErrPasswordMismatch.Error()},
		}
	case appErr.ErrInvalidRequestBody:
		return ErrorMapping{
			Status: http.StatusBadRequest,
			Error:  APIError{Code: "INVALID_REQUEST_BODY", Message: appErr.ErrInvalidRequestBody.Error()},
		}
	case appErr.ErrUnauthorized:
		return ErrorMapping{
			Status: http.StatusUnauthorized,
			Error:  APIError{Code: "UNAUTHORIZED", Message: appErr.ErrUnauthorized.Error()},
		}
	case appErr.ErrMissingSession:
		return ErrorMapping{
			Status: http.StatusUnauthorized,
			Error:  APIError{Code: "MISSING_SESSION", Message: appErr.ErrMissingSession.Error()},
		}
	case appErr.ErrInvalidCredentials:
		return ErrorMapping{
			Status: http.StatusUnauthorized,
			Error:  APIError{Code: "INVALID_CREDENTIALS", Message: appErr.ErrInvalidCredentials.Error()},
		}

	default:
		return ErrorMapping{
			Status: 500,
			Error:  APIError{Code: "internal_server_error", Message: "internal server error"},
		}
	}
}
