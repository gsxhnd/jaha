package middleware

import (
	"github.com/gsxhnd/jaha/utils"
)

type Middleware interface {
	// RequestLog(ctx *fiber.Ctx) error
}
type middleware struct {
	logger utils.Logger
}

func NewMiddleware(l utils.Logger) Middleware {
	return &middleware{
		logger: l,
	}
}
