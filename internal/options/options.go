package options

import (
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

type SendMessageOption func(*tg.MessagesSendMessageRequest)

func WithText(text string) SendMessageOption {
	return func(r *tg.MessagesSendMessageRequest) {
		r.Message = text
	}
}

func WithBuilder(eb *entity.Builder) SendMessageOption {
	return func(r *tg.MessagesSendMessageRequest) {
		r.Message, r.Entities = eb.Complete()
	}
}

func WithWebpage() SendMessageOption {
	return func(r *tg.MessagesSendMessageRequest) {
		r.SetNoWebpage(false)
	}
}

func WithEntities(e []tg.MessageEntityClass) SendMessageOption {
	return func(r *tg.MessagesSendMessageRequest) {
		r.Entities = e
	}
}

func WithMarkup(m tg.ReplyMarkupClass) SendMessageOption {
	return func(r *tg.MessagesSendMessageRequest) {
		r.ReplyMarkup = m
	}
}
