package guard

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"
	"activity-bot/internal/logger"
	"activity-bot/internal/model"
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type LevelMemberService interface {
	GetChatMember(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error)
	GetCommandLevels(ctx context.Context, chatID int64) (map[string]int16, error)
}

type CommandProvider interface {
	RegisteredCommands() map[string]*cmd.Command
}

type LevelGuard struct {
	adminService    *admin.Service
	memberService   LevelMemberService
	sessionService  cmd.SessionService
	commandProvider CommandProvider
	level           int16
}

func NewLevelGuard(
	adminService *admin.Service,
	memberService LevelMemberService,
	sessionService cmd.SessionService,
	commandProvider CommandProvider,
	level int16,
) cmd.Guard {
	return &LevelGuard{
		adminService:    adminService,
		memberService:   memberService,
		sessionService:  sessionService,
		commandProvider: commandProvider,
		level:           level,
	}
}

func (g *LevelGuard) Check(ctx *ext.Context, commandName string, stdCtx context.Context) (bool, string) {
	userID := ctx.EffectiveUser.Id

	chatID, err := cmd.GetChatID(g.sessionService, ctx, stdCtx)
	if err != nil {
		return false, "Не удалось определить чат"
	}

	m, err := g.memberService.GetChatMember(stdCtx, chatID, userID)
	if err != nil {
		logger.L.Error("Failed to get member", "error", err)
		return false, ""
	}
	userLevel := m.Level

	devLevel, _ := g.adminService.GetDevLevel(stdCtx, chatID, userID)
	if devLevel > userLevel {
		userLevel = devLevel
	}

	reqLevel := g.level
	if reqLevel == 0 && g.commandProvider != nil {
		c, ok := g.commandProvider.RegisteredCommands()[commandName]
		if !ok {
			return true, "" // Command not found in factory, allow (should not happen)
		}

		reqLevel = c.Level()

		overrides, err := g.memberService.GetCommandLevels(stdCtx, chatID)
		if err == nil {
			if lvl, ok := overrides[c.ID()]; ok {
				reqLevel = lvl
			}
		}
	}

	if userLevel < reqLevel {
		return false, "У вас недостаточно прав для этой команды."
	}

	return true, ""
}
