package cmd

import "github.com/PaulSonOfLars/gotgbot/v2/ext"

type Guard interface {
	Check(ctx *ext.Context, command string) (bool, string)
}

type GuardFunc func(ctx *ext.Context, command string) (bool, string)

func (f GuardFunc) Check(ctx *ext.Context, command string) (bool, string) {
	return f(ctx, command)
}
