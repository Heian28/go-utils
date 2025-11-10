package goerror

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// GoFiberErrorHttp is a struct that represents the error http for the GoFiberErrorCommon.
// It is used to represent the error http for the GoFiberErrorCommon.
type GoFiberErrorHttp struct {
	GoFiberErrorCommon
	StatusCode int    `json:"-"`
	StatusName string `json:"type"`
}

// Error is a function that returns the error message.
// It takes a pointer to a GoFiberErrorHttp and returns a string.
// This is used to return the error message.
func (e *GoFiberErrorHttp) Error() string {
	return e.ClientMessage
}

// GetStatusCode is a function that returns the status code.
// It takes a pointer to a GoFiberErrorCommon and returns an int.
// This is used to return the status code.
func (c *GoFiberErrorCommon) GetStatusCode() int {
	if errDict.httpCodes[c.ErrorCode] == 0 {
		return fiber.StatusInternalServerError
	}
	return errDict.httpCodes[c.ErrorCode]
}

// ToGoFiberErrorHttp is a function that converts the GoFiberErrorCommon to a GoFiberErrorHttp.
// It takes a pointer to a GoFiberErrorCommon and returns a GoFiberErrorHttp.
// This is used to convert the GoFiberErrorCommon to a GoFiberErrorHttp.
func (c *GoFiberErrorCommon) ToGoFiberErrorHttp() GoFiberErrorHttp {
	status := c.GetStatusCode()
	return GoFiberErrorHttp{
		GoFiberErrorCommon: *c,
		StatusCode:         status,
		StatusName:         getHttpStatusName(status),
	}
}

// getHttpStatusName is a function that returns the http status name.
// It takes an int and returns a string.
// This is used to return the http status name.
func getHttpStatusName(status int) string {
	if text := http.StatusText(status); text != "" {
		return text
	}
	return "INTERNAL_SERVER_ERROR"
}
