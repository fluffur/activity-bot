package conversation

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/ext"
)

type Handler interface {
	CheckUpdate(ctx *ext.Context, update *ext.Update) error
}

type Conversation struct {
	EntryPoints []Handler
	States      map[string][]Handler
	Exits       []Handler
	Storage     Storage
	TTL         time.Duration
}

func NewConversation(
	entryPoints []Handler,
	states map[string][]Handler,
	storage Storage,
	opts ...Option,
) *Conversation {
	c := &Conversation{
		EntryPoints: entryPoints,
		States:      states,
		Storage:     storage,
		TTL:         time.Hour * 24, // Default TTL
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Option func(*Conversation)

func WithExits(exits ...Handler) Option {
	return func(c *Conversation) {
		c.Exits = exits
	}
}

func WithTTL(ttl time.Duration) Option {
	return func(c *Conversation) {
		c.TTL = ttl
	}
}

func (c *Conversation) CheckUpdate(ctx *ext.Context, u *ext.Update) error {
	chatID := u.EffectiveChat().GetID()
	userID := u.EffectiveUser().GetID()

	state, err := c.Storage.Get(ctx.Context, chatID, userID)
	if err != nil {
		return err
	}

	if state != "" {
		for _, exit := range c.Exits {
			if err := exit.CheckUpdate(ctx, u); err == nil || !errors.Is(err, dispatcher.ContinueGroups) {
				// If handler matched (didn't return Skip/Continue)
				_ = c.Storage.Delete(ctx.Context, chatID, userID)
				return err
			}
		}

		handlers, ok := c.States[state]
		log.Println("hanlders ok", handlers, ok)
		if ok {
			for _, h := range handlers {
				if err := h.CheckUpdate(ctx, u); err == nil || !errors.Is(err, dispatcher.ContinueGroups) {
					return err
				}
			}
		}
	}

	// CheckUpdate entry points
	for _, entry := range c.EntryPoints {
		if err := entry.CheckUpdate(ctx, u); err == nil || err != dispatcher.ContinueGroups {
			return err
		}
	}

	return dispatcher.ContinueGroups
}

// Helpers to transition states
func SetState(ctx context.Context, storage Storage, chatID, userID int64, state string, ttl time.Duration) error {
	return storage.Set(ctx, chatID, userID, state, ttl)
}

func StopConversation(ctx context.Context, storage Storage, chatID, userID int64) error {
	return storage.Delete(ctx, chatID, userID)
}
