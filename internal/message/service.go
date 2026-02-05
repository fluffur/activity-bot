package message

import (
	"activity-bot/internal/model"
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo}
}

func (s *Service) Save(ctx context.Context, chatID int64, userID int64) error {
	return s.repo.Save(ctx, model.NewMessage(chatID, userID))
}

//
//func (s *Service) ProcessLadder(chatID, userID int64, ttl time.Duration, maxLadder int32) (int64, bool, error) {
//	ctx := context.Background()
//
//	count, sameUser, err := s.ladderRepo.Inc(ctx, chatID, userID, ttl)
//	if err != nil {
//		return 0, false, err
//	}
//
//	if maxLadder > 0 && count > int64(maxLadder) {
//
//		if err := s.ladderRepo.Reset(ctx, chatID); err != nil {
//			return count, false, err
//		}
//
//		if sameUser {
//			return count, true, nil
//		}
//
//		return count, false, nil
//	}
//
//	return count, false, nil
//}
