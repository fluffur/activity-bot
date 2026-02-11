package cmd

import (
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func Test_cleanMessage(t *testing.T) {
	tests := []struct {
		name         string
		msg          *gotgbot.Message
		wantText     string
		wantEntities int
	}{
		{
			name: "No entities",
			msg: &gotgbot.Message{
				Text: "Hello world",
			},
			wantText:     "Hello world",
			wantEntities: 0,
		},
		{
			name: "With mention at start",
			msg: &gotgbot.Message{
				Text: "@bot hello",
				Entities: []gotgbot.MessageEntity{
					{Type: "mention", Offset: 0, Length: 4},
				},
			},
			wantText:     "hello",
			wantEntities: 1,
		},
		{
			name: "With mention in middle",
			msg: &gotgbot.Message{
				Text: "hello @bot world",
				Entities: []gotgbot.MessageEntity{
					{Type: "mention", Offset: 6, Length: 4},
				},
			},
			wantText:     "hello  world",
			wantEntities: 1,
		},
		{
			name: "With URL",
			msg: &gotgbot.Message{
				Text: "/start https://google.com",
				Entities: []gotgbot.MessageEntity{
					{Type: "bot_command", Offset: 0, Length: 6},
					{Type: "url", Offset: 7, Length: 18},
				},
			},
			wantText:     "/start",
			wantEntities: 1,
		},
		{
			name: "Mention and Text Mention",
			msg: &gotgbot.Message{
				Text: "@user1 hello user2",
				Entities: []gotgbot.MessageEntity{
					{Type: "mention", Offset: 0, Length: 6},
					{Type: "text_mention", Offset: 13, Length: 5, User: &gotgbot.User{Id: 123, FirstName: "User2"}},
				},
			},
			wantText:     "hello",
			wantEntities: 2,
		},
		{
			name: "Unicode characters",
			msg: &gotgbot.Message{
				Text: "Привет @bot мир",
				Entities: []gotgbot.MessageEntity{
					{Type: "mention", Offset: 7, Length: 4},
				},
			},
			wantText:     "Привет  мир",
			wantEntities: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText, gotEntities := cleanMessage(tt.msg)
			if gotText != tt.wantText {
				t.Errorf("cleanMessage() text = %q, want %q", gotText, tt.wantText)
			}
			if len(gotEntities) != tt.wantEntities {
				t.Errorf("cleanMessage() entities count = %d, want %d", len(gotEntities), tt.wantEntities)
			}
		})
	}
}

func Test_runeIndex(t *testing.T) {

	tests := []struct {
		name      string
		text      string
		offset    int
		length    int
		wantStart int
		wantEnd   int
	}{
		{
			name:      "ASCII",
			text:      "Hello",
			offset:    1,
			length:    2,
			wantStart: 1,
			wantEnd:   3,
		},
		{
			name:      "Cyrillic - BMP",
			text:      "Привет",
			offset:    1,
			length:    1,
			wantStart: 1,
			wantEnd:   2,
		},
		{
			name:      "Emoji - Non-BMP",
			text:      "👋 World",
			offset:    0,
			length:    2,
			wantStart: 0,
			wantEnd:   1,
		},
		{
			name:      "Text after Emoji",
			text:      "👋 World",
			offset:    2,
			length:    1,
			wantStart: 1,
			wantEnd:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := runeIndex(tt.text, tt.offset, tt.length)
			if gotStart != tt.wantStart {
				t.Errorf("runeIndex() start = %d, want %d", gotStart, tt.wantStart)
			}
			if gotEnd != tt.wantEnd {
				t.Errorf("runeIndex() end = %d, want %d", gotEnd, tt.wantEnd)
			}
		})
	}
}
