package bot

import (
	"activity-bot/internal/member"

	"github.com/celestix/gotgproto/ext"
)

type memberLeftHandler struct {
	callback func(*ext.Context, *ext.Update) error
}

func (h memberLeftHandler) CheckUpdate(ctx *ext.Context, u *ext.Update) error {
	if _, _, ok := member.LeaveFromUpdate(u); !ok {
		return nil
	}
	return h.callback(ctx, u)
}
