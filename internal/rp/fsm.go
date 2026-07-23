package rp

type State string

const (
	StateIdle        State = ""
	StateAwaitGender State = "await_gender"
)

type StateData struct {
	UserID    int64
	TargetID  int64
	MessageID int
	CommandID int64
	Extra     string
	Speech    string
}
