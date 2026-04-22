package marriage

import (
	"context"
	"errors"
)

type Service struct {
	repo Repository
}

var (
	ErrAlreadyMarried  = errors.New("user already married")
	ErrRequestExists   = errors.New("marriage request already exists")
	ErrNoRequest       = errors.New("marriage request not found")
	ErrInvalidTarget   = errors.New("invalid marriage target")
	ErrNoMarriage      = errors.New("marriage not found")
	ErrNotYourMarriage = errors.New("not your marriage")
)

type RequestMarriageOutcomeType int

const (
	OutcomeSelf RequestMarriageOutcomeType = iota
	OutcomeDirect
	OutcomeRequestCreated
	OutcomeAutoAccepted
)

type RequestMarriageOutcome struct {
	Type RequestMarriageOutcomeType
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) HandleMarriageRequest(
	ctx context.Context,
	chatID, fromUserID, toUserID int64,
	isBotTarget bool,
) (RequestMarriageOutcome, error) {

	if fromUserID == toUserID {
		if _, err := s.MarryDirect(ctx, chatID, fromUserID, toUserID); err != nil {
			return RequestMarriageOutcome{}, err
		}
		return RequestMarriageOutcome{Type: OutcomeSelf}, nil
	}

	if isBotTarget {
		if _, err := s.MarryDirect(ctx, chatID, fromUserID, toUserID); err != nil {
			return RequestMarriageOutcome{}, err
		}
		return RequestMarriageOutcome{Type: OutcomeDirect}, nil
	}

	if m, err := s.repo.GetActiveMarriage(ctx, chatID, fromUserID); err != nil {
		return RequestMarriageOutcome{}, err
	} else if m != nil {
		return RequestMarriageOutcome{}, ErrAlreadyMarried
	}
	if m, err := s.repo.GetActiveMarriage(ctx, chatID, toUserID); err != nil {
		return RequestMarriageOutcome{}, err
	} else if m != nil {
		return RequestMarriageOutcome{}, ErrAlreadyMarried
	}

	if req, err := s.repo.GetActiveMarriageRequest(ctx, chatID, toUserID, fromUserID); err != nil {
		return RequestMarriageOutcome{}, err
	} else if req != nil {
		if err := s.repo.UpdateMarriageRequestStatus(ctx, req.ID, chatID, RequestStatusAccepted); err != nil {
			return RequestMarriageOutcome{}, err
		}
		if _, err := s.repo.CreateMarriage(ctx, chatID, fromUserID, toUserID); err != nil {
			return RequestMarriageOutcome{}, err
		}
		return RequestMarriageOutcome{Type: OutcomeAutoAccepted}, nil
	}

	if req, err := s.repo.GetActiveMarriageRequest(ctx, chatID, fromUserID, toUserID); err != nil {
		return RequestMarriageOutcome{}, err
	} else if req != nil {
		return RequestMarriageOutcome{}, ErrRequestExists
	}

	if _, err := s.repo.CreateMarriageRequest(ctx, chatID, fromUserID, toUserID); err != nil {
		return RequestMarriageOutcome{}, err
	}

	return RequestMarriageOutcome{Type: OutcomeRequestCreated}, nil
}

func (s *Service) GetMarriage(ctx context.Context, chatID, userID int64) (*Marriage, error) {
	return s.repo.GetActiveMarriage(ctx, chatID, userID)
}

func (s *Service) MarryDirect(ctx context.Context, chatID, user1ID, user2ID int64) (*Marriage, error) {
	if m, err := s.repo.GetActiveMarriage(ctx, chatID, user1ID); err != nil {
		return nil, err
	} else if m != nil {
		return nil, ErrAlreadyMarried
	}
	if user1ID != user2ID {
		if m, err := s.repo.GetActiveMarriage(ctx, chatID, user2ID); err != nil {
			return nil, err
		} else if m != nil {
			return nil, ErrAlreadyMarried
		}
	}
	m, err := s.repo.CreateMarriage(ctx, chatID, user1ID, user2ID)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Service) AcceptMarriageRequest(ctx context.Context, chatID, fromUserID, toUserID int64) error {
	if m, err := s.repo.GetActiveMarriage(ctx, chatID, fromUserID); err != nil {
		return err
	} else if m != nil {
		return ErrAlreadyMarried
	}
	if m, err := s.repo.GetActiveMarriage(ctx, chatID, toUserID); err != nil {
		return err
	} else if m != nil {
		return ErrAlreadyMarried
	}

	req, err := s.repo.GetActiveMarriageRequest(ctx, chatID, fromUserID, toUserID)
	if err != nil {
		return err
	}
	if req == nil {
		return ErrNoRequest
	}

	if err := s.repo.UpdateMarriageRequestStatus(ctx, req.ID, chatID, RequestStatusAccepted); err != nil {
		return err
	}
	_, err = s.repo.CreateMarriage(ctx, chatID, fromUserID, toUserID)
	return err
}

func (s *Service) RejectMarriageRequest(ctx context.Context, chatID, fromUserID, toUserID int64, cancelledByRequester bool) error {
	req, err := s.repo.GetActiveMarriageRequest(ctx, chatID, fromUserID, toUserID)
	if err != nil {
		return err
	}
	if req == nil {
		return ErrNoRequest
	}

	status := RequestStatusRejected
	if cancelledByRequester {
		status = RequestStatusCancelled
	}
	return s.repo.UpdateMarriageRequestStatus(ctx, req.ID, chatID, status)
}

func (s *Service) Divorce(ctx context.Context, chatID, actorUserID, partnerUserID int64) (*Marriage, error) {
	m, err := s.repo.GetActiveMarriage(ctx, chatID, actorUserID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrNoMarriage
	}

	if partnerUserID != 0 && m.User1.User.ID != partnerUserID && m.User2.User.ID != partnerUserID {
		return nil, ErrNotYourMarriage
	}

	if err := s.repo.DivorceMarriage(ctx, m.ID, chatID); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) ListMarriages(ctx context.Context, chatID int64) ([]Marriage, error) {
	return s.repo.ListActiveMarriages(ctx, chatID)
}
