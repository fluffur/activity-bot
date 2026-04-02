package command

import "github.com/PaulSonOfLars/gotgbot/v2"

// Guard checks whether an action should be allowed.
// Unlike cmd.Guard it receives the fully-populated *Context so it has access
// to chat, senderChatMember, etc.
type Guard interface {
	Check(b *gotgbot.Bot, ctx *Context) (ok bool, message string)
}

type GuardFunc func(b *gotgbot.Bot, ctx *Context) (bool, string)

func (f GuardFunc) Check(b *gotgbot.Bot, ctx *Context) (bool, string) {
	return f(b, ctx)
}
