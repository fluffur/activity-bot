package cmd

import (
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Guard interface {
	Check(ctx *ext.Context, command string, stdCtx context.Context) (bool, string)
}

type GuardFunc func(ctx *ext.Context, command string, stdCtx context.Context) (bool, string)

func (f GuardFunc) Check(ctx *ext.Context, command string, stdCtx context.Context) (bool, string) {
	return f(ctx, command, stdCtx)
}
