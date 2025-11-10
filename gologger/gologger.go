package gologger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logsRootDir  = "logs"
	logsFileName = "app"
)

var Logger *logrus.Logger

// config represents the configuration for the logger.
type config struct {
	isProduction bool
	serviceName  string
	prettyPrint  bool
	fields       map[string]any
}

// Option is a function that configures the logger.
// It takes a pointer to a config and returns a function that takes a pointer to a config.
// This is used to chain the options together.
type Option func(*config)

// SetIsProduction sets the logger to production mode.
// It takes a boolean and returns an Option.
// This is used to set the logger to production mode.
func SetIsProduction(isProduction bool) Option {
	return func(c *config) {
		c.isProduction = isProduction
	}
}

// SetServiceName sets the service name for the logger.
// It takes a string and returns an Option.
// This is used to set the service name for the logger.
func SetServiceName(serviceName string) Option {
	return func(c *config) {
		c.serviceName = serviceName
	}
}

// SetPrettyPrint sets the pretty print for the logger.
// It takes a boolean and returns an Option.
// This is used to set the pretty print for the logger.
func SetPrettyPrint(prettyPrint bool) Option {
	return func(c *config) {
		c.prettyPrint = prettyPrint
	}
}

// SetFields sets the fields for the logger.
// It takes a map of strings and any and returns an Option.
// This is used to set the fields for the logger.
func SetFields(fields map[string]any) Option {
	return func(c *config) {
		c.fields = fields
	}
}

// LogFileHook is a hook that logs to a file.
// It implements the logrus.Hook interface.
// It is used to log to a file.
type LogFileHook struct {
	fields    map[string]any
	Writer    io.Writer
	Formatter logrus.Formatter
}

// Levels returns all the logrus levels.
// It implements the logrus.Hook interface.
// It is used to return all the logrus levels.
func (h *LogFileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire logs the entry to the file.
// It implements the logrus.Hook interface.
// It is used to log the entry to the file.
// It returns nil if the logging is successful.
func (h *LogFileHook) Fire(entry *logrus.Entry) error {
	for k, v := range h.fields {
		entry.Data[k] = v
	}
	logBytes, err := h.Formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = h.Writer.Write(logBytes)
	if err != nil {
		return err
	}
	return nil
}

func New(opts ...Option) {
	cnf := &config{
		isProduction: false,
		serviceName:  "i-Invite Service",
		fields:       make(map[string]any),
	}
	for _, opt := range opts {
		opt(cnf)
	}

	var level logrus.Level
	if cnf.isProduction {
		level = logrus.InfoLevel
	} else {
		level = logrus.DebugLevel
	}

	log := logrus.New()
	log.SetLevel(level)

	if _, err := os.Stat(logsRootDir); os.IsNotExist(err) {
		_ = os.Mkdir(logsRootDir, 0755)
	}

	// Setup log file
	date := time.Now().Format("2006-01-02")
	logFile := &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/%s-%s.log", logsRootDir, logsFileName, date),
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	if len(cnf.serviceName) > 0 {
		cnf.fields["serviceName"] = cnf.serviceName
	}
	cnf.fields["isProduction"] = cnf.isProduction

	logFileFormatter := &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     cnf.prettyPrint,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "@level",
			logrus.FieldKeyMsg:   "@message",
			logrus.FieldKeyTime:  "@time",
		},
	}
	log.AddHook(&LogFileHook{
		Writer:    logFile,
		fields:    cnf.fields,
		Formatter: logFileFormatter,
	})
	log.SetOutput(colorable.NewColorableStdout())

	Logger = log
}
