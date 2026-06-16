package events

type Repository interface {
	SaveMessage() error
}

type Service struct {
	repository Repository
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) SaveMessage(chatID, userID int64, text string) {

}
