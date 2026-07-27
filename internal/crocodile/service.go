package crocodile

import (
	"context"
	"errors"
	"strings"
	"unicode"
)

var (
	ErrGameAlreadyExists = errors.New("game already exists")
	ErrGameNotFound      = errors.New("game not found")
	ErrNotHost           = errors.New("not game host")
)

type WordProvider interface {
	Next(ctx context.Context, chatID int64) (string, error)
}

type Service struct {
	repo     Repository
	wordRepo WordRepository
}

func NewService(
	repo Repository,
	wordRepo WordRepository,
) *Service {
	return &Service{
		repo:     repo,
		wordRepo: wordRepo,
	}
}

func (s *Service) Start(
	ctx context.Context,
	chatID int64,
	hostID int64,
) (*Game, error) {
	exists, err := s.repo.Exists(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrGameAlreadyExists
	}

	word, err := s.wordRepo.GetRandom(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.wordRepo.MarkUsed(ctx, word.ID); err != nil {
		return nil, err
	}

	game := &Game{
		ChatID: chatID,
		HostID: hostID,
		Word:   word.Word,
	}

	if err := s.repo.Create(ctx, game); err != nil {
		return nil, err
	}

	return game, nil
}

func (s *Service) Get(
	ctx context.Context,
	chatID int64,
) (*Game, error) {
	game, err := s.repo.Get(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if game == nil {
		return nil, ErrGameNotFound
	}

	return game, nil
}

func (s *Service) NextWord(
	ctx context.Context,
	chatID int64,
	hostID int64,
) (string, error) {
	game, err := s.Get(ctx, chatID)
	if err != nil {
		return "", err
	}

	if game.HostID != hostID {
		return "", ErrNotHost
	}

	word, err := s.wordRepo.GetRandom(ctx)
	if err != nil {
		return "", err
	}

	if err := s.wordRepo.MarkUsed(ctx, word.ID); err != nil {
		return "", err
	}

	if game.Word != "" {
		game.SkippedWords = append(game.SkippedWords, game.Word)
		game.SkipCount++
	}

	game.Word = word.Word

	if err := s.repo.Update(ctx, game); err != nil {
		return "", err
	}

	return word.Word, nil
}

func (s *Service) Guess(
	ctx context.Context,
	chatID int64,
	userID int64,
	text string,
) (*Game, bool, error) {
	game, err := s.Get(ctx, chatID)
	if err != nil {
		if errors.Is(err, ErrGameNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	if userID == game.HostID {
		return game, false, nil
	}

	if !EqualWord(game.Word, text) {
		return game, false, nil
	}

	if err := s.repo.Delete(ctx, chatID); err != nil {
		return nil, false, err
	}

	return game, true, nil
}

func (s *Service) Stop(
	ctx context.Context,
	chatID int64,
	hostID int64,
) error {
	game, err := s.Get(ctx, chatID)
	if err != nil {
		return err
	}

	if game.HostID != hostID {
		return ErrNotHost
	}

	return s.repo.Delete(ctx, chatID)
}

func EqualWord(a, b string) bool {
	return normalizeWord(a) == normalizeWord(b)
}

func normalizeWord(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))

	s = strings.ReplaceAll(s, "ё", "е")

	var builder strings.Builder
	builder.Grow(len(s))

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}
