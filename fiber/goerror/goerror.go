package goerror

import (
	"github.com/Heian28/go-utils/gologger"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

// NewErrorHandler is a function that handles the error.
// It takes a fiber.Ctx and an error and returns an error.
// This is used to handle the error.
func NewErrorHandler(log *logrus.Logger) func(c fiber.Ctx, err error) error {
	if log == nil {
		gologger.New(
			gologger.SetServiceName("GoError"),
		)
		log = gologger.Logger
	}
	fn := func(c fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		_e, ok := err.(*GoFiberErrorCommon)
		if ok {
			status = _e.GetStatusCode()
		} else {
			_e = ComposeClientError(UNKNOWN_ERROR, err)
		}
		httpErr := _e.ToGoFiberErrorHttp()

		gologger.CreateLogger(
			c,
			"error",
			log,
			httpErr,
			_e.ErrorMessage,
			_e.ErrorStack,
		)

		return c.Status(status).JSON(httpErr)
	}
	return fn
}
