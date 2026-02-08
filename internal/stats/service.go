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

func (s *Service) GetAllMembersStats(ctx context.Context, chatID int64, from, to *time.Time) ([]model.MessageReportMember, error) {
	return s.repo.GetReport(ctx, chatID, from, to)
}

func (s *Service) GetMemberStats(ctx context.Context, chatID, userID int64) (model.MemberStats, error) {
	return s.repo.GetReportOne(ctx, chatID, userID)
}

func (s *Service) GetChatActivityGraph(ctx context.Context, chatID int64, from, to *time.Time) (*bytes.Buffer, error) {
	activity, err := s.repo.GetMessageActivityByDayAll(ctx, chatID, from, to)
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
	end := activity[len(activity)-1].Date.Truncate(24 * time.Hour)

	if from == nil || to == nil {
		if end.Sub(start).Hours()/24 > 30 {
			start = end.AddDate(0, 0, -29)
		}
	}

	var values []chart.Value
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		values = append(values, chart.Value{
			Value: float64(activityMap[key]),
			Label: d.Format("02.01"),
		})
	}

	return s.renderBarChart("Активность чата", values, maximum)
}

func (s *Service) GetMessageActivityGraph(ctx context.Context, chatID, userID int64) (*bytes.Buffer, error) {
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

	// Limit user activity graph to last 30 days as well for consistency
	if today.Sub(start).Hours()/24 > 30 {
		start = today.AddDate(0, 0, -29)
	}

	var values []chart.Value
	for d := start; !d.After(today); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		values = append(values, chart.Value{
			Value: float64(activityMap[key]),
			Label: d.Format("02.01"),
		})
	}

	return s.renderBarChart("Статистика активности", values, maximum)
}

func (s *Service) renderBarChart(title string, values []chart.Value, maximum float64) (*bytes.Buffer, error) {
	width := 1024
	height := 512

	count := len(values)
	barWidth := 20.0
	if count > 0 {
		barWidth = float64(width) * 0.8 / float64(count)
		if barWidth > 80 {
			barWidth = 80
		}
		if barWidth < 5 {
			barWidth = 5
		}
	}

	if maximum <= 0 {
		maximum = 10
	}
	maximum = roundUpNice(maximum * 1.1)

	if count > 15 {
		step := count / 10
		if step < 2 {
			step = 2
		}
		for i := range values {
			if i%step != 0 && i != count-1 {
				values[i].Label = ""
			}
		}
	}

	graph := chart.BarChart{
		Title: title,
		TitleStyle: chart.Style{
			FontColor: drawing.ColorFromHex("333333"),
			FontSize:  16,
		},
		Width:      width,
		Height:     height,
		BarWidth:   int(barWidth),
		BarSpacing: int(barWidth * 0.3),
		Bars:       values,
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40, Bottom: 20, Left: 20, Right: 20,
			},
		},
		Canvas: chart.Style{
			FillColor: drawing.ColorFromHex("fefefe"),
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				StrokeWidth: 1.0,
				StrokeColor: drawing.ColorFromHex("e0e0e0"),
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
		XAxis: chart.Style{
			StrokeColor: drawing.ColorFromHex("e0e0e0"),
			FontSize:    10,
			FontColor:   drawing.ColorFromHex("666666"),
		},
	}

	primaryColor := drawing.ColorFromHex("3399FF")
	for i := range graph.Bars {
		graph.Bars[i].Style = chart.Style{
			FillColor:   primaryColor,
			StrokeColor: primaryColor,
			StrokeWidth: 1,
		}
	}

	buf := bytes.NewBuffer(nil)
	if err := graph.Render(chart.PNG, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func (s *Service) GetInactiveMembers(ctx context.Context, chatID int64) ([]model.InactiveMember, error) {
	return s.repo.GetInactiveMembers(ctx, chatID)
}

type ReportPeriod string

const (
	PeriodWeek  ReportPeriod = "week"
	PeriodMonth ReportPeriod = "month"
	PeriodAll   ReportPeriod = "all"
)

func ResolvePeriod(period ReportPeriod, now time.Time, weekStartDay int16) (from *time.Time, to *time.Time) {
	switch period {

	case PeriodWeek:
		isoWeekday := int(now.Weekday())
		if isoWeekday == 0 {
			isoWeekday = 7
		}

		diff := (isoWeekday - int(weekStartDay) + 7) % 7

		start := now.AddDate(0, 0, -diff)
		start = time.Date(
			start.Year(),
			start.Month(),
			start.Day(),
			0, 0, 0, 0,
			start.Location(),
		)

		end := start.AddDate(0, 0, 7).Add(-time.Second)

		return &start, &end

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
