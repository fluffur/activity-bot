package norm

type Norm struct {
	ID     int64
	ChatID int64
	Name   string
	Value  int32

	UserIDs []int64
}
