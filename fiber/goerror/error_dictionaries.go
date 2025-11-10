package goerror

import "github.com/gofiber/fiber/v3"

// UNKNOWN_ERROR is a constant that represents the unknown error code.
// It is used to represent the unknown error code.
var UNKNOWN_ERROR GoFiberErrorCode = 0

// GoFiberErrorDictionary is a struct that represents the error dictionary for the GoFiberErrorCommon.
// It is used to represent the error dictionary for the GoFiberErrorCommon.
type GoFiberErrorDictionary struct {
	errorCodes map[GoFiberErrorCode]*GoFiberErrorCommon
	httpCodes  map[GoFiberErrorCode]int
}

// errDict is a variable that represents the error dictionary for the GoFiberErrorCommon.
// It is used to represent the error dictionary for the GoFiberErrorCommon.
var errDict *GoFiberErrorDictionary = &GoFiberErrorDictionary{
	errorCodes: make(map[GoFiberErrorCode]*GoFiberErrorCommon),
	httpCodes:  make(map[GoFiberErrorCode]int),
}

// RegisterGoFiberError is a function that registers the error dictionary for the GoFiberErrorCommon.
// It takes a map of GoFiberErrorCode and a map of GoFiberErrorCode and returns nothing.
// This is used to register the error dictionary for the GoFiberErrorCommon.
func RegisterGoFiberError(
	errCodes map[GoFiberErrorCode]*GoFiberErrorCommon,
	httpCodes map[GoFiberErrorCode]int,
) {
	errDict = &GoFiberErrorDictionary{
		errorCodes: errCodes,
		httpCodes:  httpCodes,
	}

	errDict.errorCodes[UNKNOWN_ERROR] = &GoFiberErrorCommon{
		ClientMessage: "An unknown error occurred",
		ServerMessage: nil,
		ErrorCode:     UNKNOWN_ERROR,
	}

	errDict.httpCodes[UNKNOWN_ERROR] = fiber.StatusInternalServerError
}
