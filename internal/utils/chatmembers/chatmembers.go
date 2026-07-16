package chatmembers

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/emoji"
	"activity-bot/internal/permission"
	"activity-bot/internal/user"

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

		cm.User.Emojis = emoji.Random()
		result = append(result, cm)
	}

	return result
}

func fillUser(u botapi.User) user.User {
	return user.User{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Username:  u.Username,
		IsBot:     u.IsBot,
	}
}
