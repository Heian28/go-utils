package gologger

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

// HttpLogger is the struct that will be used to log http request
type HttpLogger struct {
	LogType     string              `json:"logType"`
	RequestID   string              `json:"requestId"`
	IP          string              `json:"ip"`
	Method      string              `json:"method"`
	Host        string              `json:"host"`
	Path        string              `json:"path"`
	Query       *string             `json:"query"`
	FullPath    string              `json:"fullPath"`
	RequestBody *string             `json:"requestBody"`
	Headers     map[string][]string `json:"headers"`
	Time        time.Time           `json:"time"`
}

// HttpResponseLogger is the struct that will be used to log http response
type HttpResponseLogger struct {
	HttpLogger
	Status       int       `json:"status"`
	ResponseBody *string   `json:"responseBody"`
	Time         time.Time `json:"time"`
}

// HttpErrorLogger is the struct that will be used to log http error
type HttpErrorLogger struct {
	HttpLogger
	Status       int       `json:"status"`
	ResponseBody any       `json:"responseBody"`
	ErrorMessage any       `json:"errorMessage"`
	ErrorStack   any       `json:"errorStack"`
	Time         time.Time `json:"time"`
}

// CreateLogger is the function that will be used to create a logger
func CreateLogger(
	c fiber.Ctx,
	logType string,
	log *logrus.Logger,
	httpError any,
	errorMessage any,
	errorStack any,
) {
	var (
		host    string
		path    string
		ip      string
		method  string
		query   *string
		body    *string
		resBody *string
		headers map[string][]string
	)

	ip = c.IP()
	req := c.Request()
	method = c.Method()
	headers = c.GetReqHeaders()
	host = string(req.Host())
	path = string(req.URI().Path())
	reqId := c.Locals(fiber.HeaderXRequestID).(string)

	if req.URI().QueryString() != nil {
		_q := string(req.URI().QueryString())
		query = &_q
	}

	if req.Body() != nil {
		_b := string(req.Body())
		if len(_b) > 0 {
			body = &_b
		}
	}

	_log := HttpLogger{
		LogType:     "request",
		Host:        host,
		Path:        path,
		IP:          ip,
		Query:       query,
		FullPath:    req.URI().String(),
		RequestBody: body,
		Method:      method,
		RequestID:   reqId,
		Headers:     headers,
		Time:        time.Now(),
	}

	switch logType {
	case "request":
		buffLog := buildBuffer(_log)
		log.Info(buffLog)
	case "response":
		_res := c.Response()
		_resB := _res.Body()
		if _resB != nil {
			_bb := string(_resB)
			if len(_bb) > 0 {
				resBody = &_bb
			}
		}
		_resLog := HttpResponseLogger{
			HttpLogger:   _log,
			Status:       c.Response().StatusCode(),
			ResponseBody: resBody,
			Time:         time.Now(),
		}
		_resLog.LogType = "response"
		buffLog := buildBuffer(_resLog)
		log.Info(buffLog)
	case "error":
		_resLog := HttpErrorLogger{
			HttpLogger:   _log,
			Status:       c.Response().StatusCode(),
			ResponseBody: httpError,
			ErrorMessage: errorMessage,
			ErrorStack:   errorStack,
			Time:         time.Now(),
		}
		_resLog.LogType = "error"
		buffLog := buildBuffer(_resLog)
		log.Error(buffLog)
	default:
		buffLog := buildBuffer(_log)
		log.Info(buffLog)
	}
}

// buildBuffer is the function that will be used to build a buffer
func buildBuffer(log any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.Encode(log)
	return buf.String()
}
