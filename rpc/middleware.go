package rpc

import (
	"errors"
	"net"
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

// localhostOnly restricts an endpoint to requests originating from the local machine.
func localhostOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ip := net.ParseIP(c.RealIP())
		if ip == nil || !ip.IsLoopback() {
			return echo.NewHTTPError(http.StatusForbidden, "only accessible from localhost")
		}
		return next(c)
	}
}
