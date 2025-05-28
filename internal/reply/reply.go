package reply

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	error   string
	message string
}

func Err(error string, message string) ErrorResponse {
	return ErrorResponse{
		error:   error,
		message: message,
	}
}

func BadRequest(ctx echo.Context, err ErrorResponse) error {
	return ctx.JSON(
		http.StatusBadRequest,
		map[string]string{
			"error":   err.error,
			"message": err.message,
		},
	)
}
