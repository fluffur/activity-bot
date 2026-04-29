package rp

import (
	"activity-bot/internal/model"
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func normalizeTrigger(trigger string) string {
	return strings.ToLower(strings.Join(strings.Fields(trigger), " "))
}

func (s *Service) Upsert(ctx context.Context, cmd model.RPCommand) error {
	cmd.Trigger = strings.TrimSpace(cmd.Trigger)
	return s.repo.Upsert(ctx, cmd)
}

func (s *Service) Delete(ctx context.Context, chatID int64, trigger string) error {
	return s.repo.Delete(ctx, chatID, normalizeTrigger(trigger))
}

func (s *Service) ListByChat(ctx context.Context, chatID int64) ([]model.RPCommand, error) {
	return s.repo.ListByChat(ctx, chatID)
}

func (s *Service) Match(ctx context.Context, chatID int64, text string) (model.RPCommand, bool, error) {
	commands, err := s.repo.ListByChat(ctx, chatID)
	if err != nil {
		return model.RPCommand{}, false, err
	}

	normalizedText := normalizeTrigger(text)
	var best model.RPCommand
	bestLen := 0

	for _, cmd := range commands {
		candidate := normalizeTrigger(cmd.Trigger)
		if candidate == "" {
			continue
		}

		if normalizedText == candidate || strings.HasPrefix(normalizedText, candidate+" ") {
			if len([]rune(candidate)) > bestLen {
				best = cmd
				bestLen = len([]rune(candidate))
			}
		}
	}

	if bestLen == 0 {
		return model.RPCommand{}, false, nil
	}

	return best, true, nil
}
