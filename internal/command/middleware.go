package command

import (
	"errors"

	"github.com/celestix/gotgproto/ext"
)

type Middleware interface {
	CheckUpdate(ctx *Context, u *ext.Update) error
}

var ErrStop = errors.New("middleware: stop execution")
