package command

type Factory struct {
	userProvider    UserProvider
	memberProvider  ChatMemberProvider
	chatProvider    ChatProvider
	sessionService  SessionService
	defaultTriggers []string
}

func NewCommandFactory(up UserProvider, mp ChatMemberProvider, cp ChatProvider, ss SessionService, triggers ...string) *Factory {
	return &Factory{
		userProvider:    up,
		memberProvider:  mp,
		chatProvider:    cp,
		sessionService:  ss,
		defaultTriggers: triggers,
	}
}

func (f *Factory) New(name string, response Response) Command {
	return NewCommand(name, response).
		SetTriggers(f.defaultTriggers...).
		SetProviders(f.userProvider, f.memberProvider, f.chatProvider, f.sessionService)
}
