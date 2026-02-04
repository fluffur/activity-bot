package stats

import (
	"activity-bot/internal/model"
	"bytes"
	"context"
	"fmt"
	"math"
	"time"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo}
}

func (s *Service) GetAllMembersStats(chatID int64, from, to *time.Time) ([]model.MessageReportMember, error) {
	ctx := context.Background()
	return s.repo.GetReport(ctx, chatID, from, to)
}

func (s *Service) GetMemberStats(chatID, userID int64) (model.MemberStats, error) {
	ctx := context.Background()
	return s.repo.GetReportOne(ctx, chatID, userID)
}

func (s *Service) GetMessageActivityGraph(chatID, userID int64) (*bytes.Buffer, error) {
	ctx := context.Background()

	activity, err := s.repo.GetMessageActivityByDay(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}

	if len(activity) == 0 {
		return nil, nil
	}

	activityMap := make(map[string]int64, len(activity))
	for _, a := range activity {
		activityMap[a.Date.Format("2006-01-02")] = a.Count
	}

	start := activity[0].Date
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)

	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)

	buf := bytes.NewBuffer(nil)

	daysCount := int(today.Sub(start).Hours()/24) + 1
	width := daysCount * 120

	if width < 400 {
		width = 400
	}
	if width > 1000 {
		width = 1000
	}
	barWidth := width / daysCount * 65 / 100

	if barWidth < 20 {
		barWidth = 20
	}
	if barWidth > 200 {
		barWidth = 200
	}
	var maximum float64
	for _, val := range activityMap {
		if float64(val) > maximum {
			maximum = float64(val)
		}
	}

	maximum = maximum * 1.1
	maximum = roundUpNice(maximum)
	if maximum < 1 {
		maximum = 1
	}
	minHeight := 300
	pixelsPerUnit := 10.0
	extraHeight := 80

	height := int(float64(minHeight) + maximum*pixelsPerUnit + float64(extraHeight))

	if height > 800 {
		height = 800
	}
	graph := chart.BarChart{
		Title:    "Статистика активности",
		Width:    width,
		Height:   height,
		BarWidth: barWidth,
		Bars:     []chart.Value{},
		YAxis: chart.YAxis{
			Name: "Сообщения",
			NameStyle: chart.Style{
				FontSize:  10,
				FontColor: drawing.Color{A: 155},
			},
			Range: &chart.ContinuousRange{
				Min: 0,
				Max: maximum,
			},
			Ticks: []chart.Tick{
				{Value: 0},
				{Value: maximum * 0.25},
				{Value: maximum * 0.5},
				{Value: maximum * 0.75},
				{Value: maximum},
			},
			ValueFormatter: func(v interface{}) string {
				if val, ok := v.(float64); ok {
					return fmt.Sprintf("%.0f", val)
				}
				return ""
			},
		},
	}

	for d := start; !d.After(today); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")

		graph.Bars = append(graph.Bars, chart.Value{
			Value: float64(activityMap[key]),
			Label: d.Format("02.01"),
		})

	}

	if err := graph.Render(chart.PNG, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func (s *Service) GetInactiveMembers(chatID int64) ([]model.InactiveMember, error) {
	ctx := context.Background()
	return s.repo.GetInactiveMembers(ctx, chatID)
}

type ReportPeriod string

const (
	PeriodWeek  ReportPeriod = "week"
	PeriodMonth ReportPeriod = "month"
	PeriodAll   ReportPeriod = "all"
)

func ResolvePeriod(period ReportPeriod, now time.Time) (from *time.Time, to *time.Time) {
	switch period {

	case PeriodWeek:
		weekday := int(now.Weekday())
		daysSinceMonday := (weekday + 6) % 7
		monday := now.AddDate(0, 0, -daysSinceMonday)
		monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
		sunday := monday.AddDate(0, 0, 6)
		sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, sunday.Location())

		return &monday, &sunday

	case PeriodMonth:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, 0).Add(-time.Second)
		return &start, &end

	case PeriodAll:
		return nil, nil
	}

	return nil, nil
}

func roundUpNice(v float64) float64 {
	if v <= 10 {
		return 10
	}
	if v <= 20 {
		return 20
	}
	if v <= 50 {
		return 50
	}
	if v <= 100 {
		return 100
	}

	pow := math.Pow(10, math.Floor(math.Log10(v)))
	return math.Ceil(v/pow) * pow
}
