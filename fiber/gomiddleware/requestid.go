package gomiddleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// XRequestID is a constant that represents the request id header.
// It is used to represent the request id header.
const XRequestID string = "X-Request-ID"

// RequestID is a function that returns a fiber.Handler.
// It is used to return a fiber.Handler.
func RequestID() fiber.Handler {
	fn := func(c fiber.Ctx) error {
		rqId := c.Get(XRequestID, "")
		if len(rqId) == 0 {
			rqId = uuid.New().String()
		}
		c.Locals(XRequestID, rqId)
		c.Set(XRequestID, rqId)
		return c.Next()
	}
	return fn
}
