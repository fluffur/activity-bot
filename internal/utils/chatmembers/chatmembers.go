package chatmembers

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/permission"
	"activity-bot/internal/user"
	"time"

	"github.com/gotd/botapi"
)

func ExtractMembers(members []botapi.ChatMember) []chatmember.ChatMember {
	var result []chatmember.ChatMember

	for _, member := range members {
		cm := chatmember.ChatMember{}

		switch m := member.(type) {
		case *botapi.ChatMemberMember:
			cm.User = fillUser(m.User)
			cm.Tag = m.Tag

		case *botapi.ChatMemberAdministrator:
			cm.User = fillUser(m.User)
			cm.Tag = m.CustomTitle

		case *botapi.ChatMemberOwner:
			cm.User = fillUser(m.User)
			cm.Tag = m.CustomTitle
			cm.Status = permission.StatusOwner

		case *botapi.ChatMemberRestricted:
			cm.User = fillUser(m.User)
			cm.Tag = m.Tag

		case *botapi.ChatMemberBanned:
			cm.User = fillUser(m.User)

		case *botapi.ChatMemberLeft:
			cm.User = fillUser(m.User)
		}

		result = append(result, cm)
	}

	return result
}

func fillUser(u botapi.User) user.User {
	return user.New(
		u.ID,
		u.FirstName,
		u.LastName,
		u.Username,
		user.GenderUnknown,
		u.IsBot,
		time.Now(),
	)
}
