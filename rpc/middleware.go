package rpc

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

func customHTTPErrorHandler(err error, c echo.Context) {
	var (
		code    = http.StatusInternalServerError
		message = "internal server error"
	)

	var he *echo.HTTPError
	if errors.As(err, &he) {
		code = he.Code
		if msg, ok := he.Message.(string); ok {
			message = msg
		}
	}

	if code >= 500 {
		c.Logger().Error(err)
	}

	if c.Response().Committed {
		return
	}

	c.JSON(code, ErrorResponse{Error: message})
}
