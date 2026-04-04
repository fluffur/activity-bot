package command

type Factory struct {
	userProvider    UserProvider
	memberProvider  ChatMemberProvider
	chatProvider    ChatProvider
	sessionService  SessionService
	defaultTriggers []string
	commands        []*Command
	devID           int64
}

func NewCommandFactory(up UserProvider, mp ChatMemberProvider, cp ChatProvider, ss SessionService, devID int64, triggers ...string) *Factory {
	return &Factory{
		userProvider:    up,
		memberProvider:  mp,
		chatProvider:    cp,
		sessionService:  ss,
		defaultTriggers: triggers,
		devID:           devID,
	}
}

func (f *Factory) New(name string, response Response) *Command {
	cmd := NewCommand(name, response).
		SetDevID(f.devID).
		SetPrefixes(f.defaultTriggers...).
		SetProviders(f.userProvider, f.memberProvider, f.chatProvider, f.sessionService)
	f.commands = append(f.commands, cmd)

	return cmd
}

func (f *Factory) ConfigurableCommands() []*Command {
	var res []*Command
	for _, c := range f.commands {
		if c.description != "" {
			res = append(res, c)
		}
	}
	return res
}
