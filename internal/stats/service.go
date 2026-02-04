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
	var maximum float64
	for _, a := range activity {
		activityMap[a.Date.Format("2006-01-02")] = a.Count
		if float64(a.Count) > maximum {
			maximum = float64(a.Count)
		}
	}

	start := activity[0].Date.Truncate(24 * time.Hour)
	today := time.Now().Truncate(24 * time.Hour)

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

	if maximum > 0 {
		maximum = maximum * 1.1
	}
	if maximum >= 50 {
		maximum = roundUpNice(maximum)
	}
	if maximum < 1 {
		maximum = 1
	}

	maxGraphHeight := 500.0
	pixelsPerUnit := maxGraphHeight / maximum
	if pixelsPerUnit < 5 {
		pixelsPerUnit = 5
	}
	height := int(maximum*pixelsPerUnit + 100)
	if height > 800 {
		height = 800
	}

	buf := bytes.NewBuffer(nil)

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
			Ticks: buildNiceTicks(maximum),
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
func buildNiceTicks(max float64) []chart.Tick {
	if max <= 0 {
		return []chart.Tick{{Value: 0, Label: "0"}}
	}

	steps := []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000}
	var step float64

	for _, s := range steps {
		if math.Ceil(max/s) <= 6 {
			step = s
			break
		}
	}
	if step == 0 {
		step = math.Ceil(max / 5)
	}

	maxTick := math.Floor(max/step) * step
	if maxTick < max {
		maxTick += step
	}

	var ticks []chart.Tick
	for v := 0.0; v <= maxTick; v += step {
		ticks = append(ticks, chart.Tick{
			Value: v,
			Label: fmt.Sprintf("%.0f", v),
		})
	}

	return ticks
}
