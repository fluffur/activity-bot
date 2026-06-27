package norm

type Norm struct {
	ID     int64
	ChatID int64
	Name   string
	Value  int32

	UserIDs []int64
}

func (n Norm) BelongsToUser(userID int64) bool {
	if len(n.UserIDs) == 0 {
		return true
	}

	for _, id := range n.UserIDs {
		if id == userID {
			return true
		}
	}

	return false
}
