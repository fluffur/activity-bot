package cmd

import (
	"testing"
)

func TestCommand_matchCommand(t *testing.T) {
	tests := []struct {
		name        string
		command     *Command
		botUsername string
		text        string
		wantRest    string
		wantMatched bool
	}{
		{
			name: "Simple match",
			command: &Command{
				commands: []string{"start"},
				triggers: []string{"/"},
			},
			botUsername: "Bot",
			text:        "/start",
			wantRest:    "",
			wantMatched: true,
		},
		{
			name: "Match with args",
			command: &Command{
				commands: []string{"start"},
				triggers: []string{"/"},
			},
			botUsername: "Bot",
			text:        "/start arg1 arg2",
			wantRest:    "arg1 arg2",
			wantMatched: true,
		},
		{
			name: "Match with bot username",
			command: &Command{
				commands: []string{"start"},
				triggers: []string{"/"},
			},
			botUsername: "MyBot",
			text:        "/start@MyBot",
			wantRest:    "",
			wantMatched: true,
		},
		{
			name: "Match with bot username and args",
			command: &Command{
				commands: []string{"start"},
				triggers: []string{"/"},
			},
			botUsername: "MyBot",
			text:        "/start@MyBot arg1",
			wantRest:    "arg1",
			wantMatched: true,
		},
		{
			name: "Case insensitive",
			command: &Command{
				commands: []string{"start"},
				triggers: []string{"/"},
			},
			botUsername: "Bot",
			text:        "/START",
			wantRest:    "",
			wantMatched: true,
		},
		{
			name: "Wrong trigger",
			command: &Command{
				commands: []string{"start"},
				triggers: []string{"!"},
			},
			botUsername: "Bot",
			text:        "/start",
			wantRest:    "",
			wantMatched: false,
		},
		{
			name: "Wrong command",
			command: &Command{
				commands: []string{"start"},
				triggers: []string{"/"},
			},
			botUsername: "Bot",
			text:        "/stop",
			wantRest:    "",
			wantMatched: false,
		},
		{
			name: "Alias check",
			command: &Command{
				commands: []string{"start", "begin"},
				triggers: []string{"/"},
			},
			botUsername: "Bot",
			text:        "/begin",
			wantRest:    "",
			wantMatched: true,
		},
		{
			name: "Partial match fail",
			command: &Command{
				commands: []string{"star"},
				triggers: []string{"/"},
			},
			botUsername: "Bot",
			text:        "/start",
			wantRest:    "",
			wantMatched: false,
		},
		{
			name: "Multiple triggers",
			command: &Command{
				commands: []string{"start"},
				triggers: []string{"/", "!"},
			},
			botUsername: "Bot",
			text:        "!start",
			wantRest:    "",
			wantMatched: true,
		},
		{
			name: "Russian command",
			command: &Command{
				commands: []string{"старт"},
				triggers: []string{"/"},
			},
			botUsername: "Bot",
			text:        "/старт",
			wantRest:    "",
			wantMatched: true,
		},
		{
			name: "Russian command",
			command: &Command{
				commands: []string{"старт"},
				triggers: []string{"!"},
			},
			botUsername: "Bot",
			text:        "! старт",
			wantRest:    "",
			wantMatched: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRest, gotMatched := tt.command.matchCommand(tt.text, tt.botUsername)
			if gotMatched != tt.wantMatched {
				t.Errorf("matchCommand() matched = %v, want %v", gotMatched, tt.wantMatched)
			}
			if gotRest != tt.wantRest {
				t.Errorf("matchCommand() rest = %q, want %q", gotRest, tt.wantRest)
			}
		})
	}
}
