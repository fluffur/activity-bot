package helpers

import (
	"testing"
	"time"
)

func TestDateParser_Parse(t *testing.T) {
	fixedTime := time.Date(2023, time.October, 15, 12, 0, 0, 0, time.UTC)
	parser := &DateParser{now: func() time.Time { return fixedTime }}

	tests := []struct {
		name   string
		arg    string
		want   time.Time
		wantOk bool
	}{
		{
			name:   "Today",
			arg:    "сегодня",
			want:   time.Date(2023, time.October, 15, 12, 0, 0, 0, time.UTC),
			wantOk: true,
		},
		{
			name:   "Tomorrow",
			arg:    "завтра",
			want:   time.Date(2023, time.October, 16, 12, 0, 0, 0, time.UTC),
			wantOk: true,
		},
		{
			name:   "Yesterday",
			arg:    "вчера",
			want:   time.Date(2023, time.October, 14, 12, 0, 0, 0, time.UTC),
			wantOk: true,
		},
		{
			name:   "Russian Date with Year",
			arg:    "1 января 2023",
			want:   time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC),
			wantOk: true,
		},
		{
			name:   "Russian Date without Year",
			arg:    "1 января",
			want:   time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC),
			wantOk: true,
		},
		{
			name:   "Standard Date",
			arg:    "01.01.2023",
			want:   time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC),
			wantOk: true,
		},
		{
			name:   "Standard Date Short",
			arg:    "01.01",
			want:   time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC),
			wantOk: true,
		},
		{
			name:   "Days - Positive",
			arg:    "10",
			want:   time.Date(2023, time.October, 25, 12, 0, 0, 0, time.UTC),
			wantOk: true,
		},
		{
			name:   "Days - Negative text",
			arg:    "10 дней назад",
			want:   time.Date(2023, time.October, 5, 12, 0, 0, 0, time.UTC),
			wantOk: true,
		},
		{
			name:   "Invalid",
			arg:    "invalid",
			want:   time.Time{},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parser.Parse(tt.arg)
			if ok != tt.wantOk {
				t.Errorf("Parse() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && !got.Equal(tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateParser_ParseRange(t *testing.T) {
	fixedTime := time.Date(2023, time.October, 15, 12, 0, 0, 0, time.UTC)
	parser := &DateParser{now: func() time.Time { return fixedTime }}

	firstJan := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
	tenthJan := time.Date(2023, time.January, 10, 12, 0, 0, 0, time.UTC).AddDate(0, 0, 1).Add(-time.Second)

	tests := []struct {
		name     string
		args     []string
		wantFrom *time.Time
		wantTo   *time.Time
		wantOk   bool
	}{
		{
			name:     "Last N days",
			args:     []string{"10"},
			wantFrom: func() *time.Time { t := fixedTime.AddDate(0, 0, -10); return &t }(),
			wantTo:   nil,
			wantOk:   true,
		},
		{
			name:     "Simple Range - Days of Month",
			args:     []string{"1-10"},
			wantFrom: func() *time.Time { t := time.Date(2023, time.October, 1, 12, 0, 0, 0, time.UTC); return &t }(),
			wantTo: func() *time.Time {
				t := time.Date(2023, time.October, 10, 12, 0, 0, 0, time.UTC).AddDate(0, 0, 1).Add(-time.Second)
				return &t
			}(),
			wantOk: true,
		},
		{
			name:     "Full Date Range",
			args:     []string{"01.01.2023-10.01.2023"},
			wantFrom: &firstJan,
			wantTo:   &tenthJan,
			wantOk:   true,
		},
		{
			name:     "Keywords Range",
			args:     []string{"с", "01.01.2023", "по", "10.01.2023"},
			wantFrom: &firstJan,
			wantTo:   &tenthJan,
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFrom, gotTo, ok := parser.ParseRange(tt.args)
			if ok != tt.wantOk {
				t.Errorf("ParseRange() ok = %v, want %v", ok, tt.wantOk)
				return
			}

			if gotFrom == nil && tt.wantFrom != nil {
				t.Errorf("ParseRange() gotFrom = nil, want %v", tt.wantFrom)
			} else if gotFrom != nil && tt.wantFrom != nil {
				if !gotFrom.Equal(*tt.wantFrom) {
					t.Errorf("ParseRange() gotFrom = %v, want %v", gotFrom, tt.wantFrom)
				}
			}

			if gotTo == nil && tt.wantTo != nil {
				t.Errorf("ParseRange() gotTo = nil, want %v", tt.wantTo)
			} else if gotTo != nil && tt.wantTo != nil {
				if !gotTo.Equal(*tt.wantTo) {
					t.Errorf("ParseRange() gotTo = %v, want %v", gotTo, tt.wantTo)
				}
			}
		})
	}
}
