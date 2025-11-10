package goerror

import "fmt"

// GoFiberErrorCode is a type that represents the error code for the GoFiberErrorCommon.
// It is used to represent the error code for the GoFiberErrorCommon.
type GoFiberErrorCode uint

// GoFiberErrorCommon is a struct that represents the error common for the GoFiberErrorCommon.
// It is used to represent the error common for the GoFiberErrorCommon.
type GoFiberErrorCommon struct {
	ClientMessage string           `json:"message"`
	ServerMessage any              `json:"data"`
	ErrorCode     GoFiberErrorCode `json:"code"`
	ErrorMessage  *string          `json:"-"`
	ErrorStack    *string          `json:"-"`
}

// Error is a function that returns the error message.
// It takes a pointer to a GoFiberErrorCommon and returns a string.
// This is used to return the error message.
func (e *GoFiberErrorCommon) Error() string {
	var (
		errMsg   string
		errStack string
	)
	if e.ErrorMessage != nil {
		errMsg = *e.ErrorMessage
	}
	if e.ErrorStack != nil {
		errStack = *e.ErrorStack
	}
	return fmt.Sprintf("Error: %s, Trace: %s", errMsg, errStack)
}

// SetClientMessage is a function that sets the client message.
// It takes a string and returns nothing.
// This is used to set the client message.
func (e *GoFiberErrorCommon) SetClientMessage(msg string) {
	e.ClientMessage = msg
}

// SetServerMessage is a function that sets the server message.
// It takes a any and returns nothing.
// This is used to set the server message.
func (e *GoFiberErrorCommon) SetServerMessage(msg any) {
	e.ServerMessage = msg
}

// ComposeClientError is a function that composes the client error.
// It takes a GoFiberErrorCode and an error and returns a pointer to a GoFiberErrorCommon.
// This is used to compose the client error.
func ComposeClientError(code GoFiberErrorCode, err error) *GoFiberErrorCommon {
	var (
		em *string
		et *string
		cm string              = "Unknown error"
		sm any                 = "Unknown error"
		oe *GoFiberErrorCommon = errDict.errorCodes[code]
	)
	if err != nil {
		ori := err.Error()
		em = &ori

		ori2 := fmt.Sprintf("%+v", err)
		et = &ori2

		if code == UNKNOWN_ERROR {
			sm = ori2
		}
	}

	if oe == nil {
		return &GoFiberErrorCommon{
			ClientMessage: cm,
			ServerMessage: sm,
			ErrorCode:     code,
			ErrorMessage:  em,
			ErrorStack:    et,
		}
	}

	if _e, ok := err.(*GoFiberErrorCommon); ok {
		_e.ErrorMessage = em
		_e.ErrorStack = et
		return _e
	}

	return &GoFiberErrorCommon{
		ClientMessage: oe.ClientMessage,
		ServerMessage: oe.ServerMessage,
		ErrorCode:     code,
		ErrorMessage:  em,
		ErrorStack:    et,
	}
}
