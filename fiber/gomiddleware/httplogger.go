package gomiddleware

import (
	"time"

	"github.com/Heian28/go-utils/gologger"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

// HttpRequestLog is a struct that represents the request log.
type HttpRequestLog struct {
	LogType   string              `json:"logType"`
	Host      string              `json:"host"`
	Path      string              `json:"path"`
	IP        string              `json:"ip"`
	Query     *string             `json:"query"`
	ReqBody   *string             `json:"reqBody"`
	Method    string              `json:"method"`
	RequestID string              `json:"requestId"`
	Headers   map[string][]string `json:"headers"`
	Time      time.Time           `json:"time"`
}

// HttpResponseLog is a struct that represents the response log.
type HttpResponseLog struct {
	HttpRequestLog
	Status  int       `json:"status"`
	ResBody *string   `json:"resBody"`
	Time    time.Time `json:"time"`
}

// HttpLogger is a function that returns a fiber.Handler.
// It is used to return a fiber.Handler.
func HttpLogger(log *logrus.Logger) fiber.Handler {
	if log == nil {
		gologger.New(
			gologger.SetServiceName("HttpLogger"),
		)
		log = gologger.Logger
	}

	fn := func(c fiber.Ctx) error {
		gologger.CreateLogger(
			c,
			"request",
			log,
			nil,
			nil,
			nil,
		)
		resErr := c.Next()
		gologger.CreateLogger(
			c,
			"response",
			log,
			nil,
			nil,
			nil,
		)
		return resErr
	}
	return fn
}
