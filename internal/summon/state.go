package summon

type State string

const (
	StateIdle              State = ""
	StateAwaitConfirmation State = "await_confirmation"
)

type StateData struct {
	Text      string
	MessageID int
	ChatID    int64
	UserID    int64
}
