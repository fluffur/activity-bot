package weather

import (
	"activity-bot/internal/action"
	"activity-bot/internal/cctx"
	"activity-bot/internal/chat/handler"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/rule"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/botapi"
)

type Handler struct {
	client *Client
}

func NewHandler(client *Client) *Handler {
	return &Handler{
		client: client,
	}
}
func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"weatherfun",
			h.WeatherFun,
			i18n.Cmd.Weather.Desc,
			handler.CategoryChat,
			option.WithAliases("погода сиксевенбург"),
		),
		action.NewCommand(
			"weather",
			h.Weather,
			i18n.Cmd.Weather.Desc,
			handler.CategoryChat,
			option.WithAliases("погода"),
			option.WithRules(rule.Text().Validate(func(s string) bool {
				return len([]rune(s)) <= 128
			})),
		),
	}
}

func (h *Handler) WeatherFun(c *botapi.Context) error {
	now := time.Now()

	w := &ForecastResponse{
		Location: Location{
			Name:    "Сиксевенбург",
			Country: "67",
		},
		Current: Current{
			TempC:      67,
			FeelsLikeC: 67,
			WindKph:    67 * 3.6,
			Humidity:   67,
			Cloud:      67,
			Condition: Condition{
				Text: "Переменная облачность",
			},
		},
	}

	w.Forecast.Forecastday = []ForecastDay{
		{
			Date: now.Format("2006-01-02"),
			Day: struct {
				MinTemp   float64   `json:"mintemp_c"`
				MaxTemp   float64   `json:"maxtemp_c"`
				Condition Condition `json:"condition"`
			}{
				MinTemp: 67,
				MaxTemp: 67,
				Condition: Condition{
					Text: "Переменная облачность",
				},
			},
			Hour: fakeHours(now),
		},
		{
			Date: now.AddDate(0, 0, 1).Format("2006-01-02"),
			Day: struct {
				MinTemp   float64   `json:"mintemp_c"`
				MaxTemp   float64   `json:"maxtemp_c"`
				Condition Condition `json:"condition"`
			}{
				MinTemp: 67,
				MaxTemp: 67,
				Condition: Condition{
					Text: "Переменная облачность",
				},
			},
			Hour: fakeHours(now.AddDate(0, 0, 1)),
		},
	}

	loc := cctx.MustLocalizer(c)

	current := loc.T(
		i18n.Cmd.Weather.Current,
		i18n.CmdWeatherCurrentData{
			Temp:      "+67°",
			Condition: "Переменная облачность",
			FeelsLike: "+67°",
			Wind:      "67",
			Humidity:  67,
		},
	)

	data := i18n.CmdWeatherMessageData{
		City:     w.Location.Name,
		Current:  current,
		Forecast: renderForecast(loc, w),
	}

	_, err := c.Reply(
		loc.T(i18n.Cmd.Weather.Message, data),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func fakeHours(date time.Time) []Hour {
	hours := make([]Hour, 0, 24)

	for i := 0; i < 24; i++ {
		hours = append(hours, Hour{
			Time: date.Format("2006-01-02") +
				fmt.Sprintf(" %02d:00", i),
			Temp: 67,
			Condition: Condition{
				Text: "Переменная облачность",
			},
		})
	}

	return hours
}

func (h *Handler) Weather(c *botapi.Context) error {
	place, ok := cctx.MustArgs(c).Text()
	if !ok {
		return nil
	}

	loc := cctx.MustLocalizer(c)

	w, err := h.client.Forecast(c, place)
	if err != nil {
		return fmt.Errorf("forecast: %w", err)
	}

	current := loc.T(
		i18n.Cmd.Weather.Current,
		i18n.CmdWeatherCurrentData{
			Temp:      fmt.Sprintf("%+.0f°", w.Current.TempC),
			Condition: w.Current.Condition.Text,
			FeelsLike: fmt.Sprintf("%+.0f°", w.Current.FeelsLikeC),
			Wind:      fmt.Sprintf("%.2f", w.Current.WindKph/3.6),
			Humidity:  w.Current.Humidity,
		},
	)

	data := i18n.CmdWeatherMessageData{
		City:     w.Location.Name,
		Current:  current,
		Forecast: renderForecast(loc, w),
	}

	_, err = c.Reply(loc.T(
		i18n.Cmd.Weather.Message,
		data,
	), botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func renderForecast(loc *i18n.Localizer, w *ForecastResponse) string {
	var b strings.Builder

	for i, day := range w.Forecast.Forecastday {
		switch i {
		case 0:
			b.WriteString("\n")
			b.WriteString(loc.T(i18n.Cmd.Weather.Today, nil))
			b.WriteString("\n")

		case 1:
			b.WriteString("\n")
			b.WriteString(loc.T(i18n.Cmd.Weather.Tomorrow, nil))
			b.WriteString("\n")

		default:
			date, _ := time.Parse("2006-01-02", day.Date)

			fmt.Fprintf(
				&b,
				"\n%s <b>%s</b> <code>+%.0f</code>..<code>+%.0f</code>°",
				weatherEmoji(day.Day.Condition.Text),
				tghtml.DateTime(date, "d", date.Format("02 Jan")),
				day.Day.MinTemp,
				day.Day.MaxTemp,
			)

			continue
		}

		periods := []struct {
			Name string
			From int
			To   int
		}{
			{"Ночью", 0, 6},
			{"Утром", 6, 12},
			{"Днём", 12, 18},
			{"Вечером", 18, 24},
		}

		for _, p := range periods {
			for _, hour := range day.Hour {
				t, _ := time.Parse(
					"2006-01-02 15:04",
					hour.Time,
				)

				if t.Hour() >= p.From && t.Hour() < p.To {
					fmt.Fprintf(
						&b,
						"%s <code>%8s:</code> <code>%+0.0f°</code>\n",
						weatherEmoji(hour.Condition.Text),
						p.Name,
						hour.Temp,
					)
					break
				}
			}
		}
	}

	return b.String()
}

func weatherEmoji(condition string) string {
	condition = strings.ToLower(condition)

	switch {
	case strings.Contains(condition, "дожд"):
		return "🌧"
	case strings.Contains(condition, "лив"):
		return "🌧"
	case strings.Contains(condition, "снег"):
		return "🌨"
	case strings.Contains(condition, "метел"):
		return "🌨"
	case strings.Contains(condition, "туман"):
		return "🌫"
	case strings.Contains(condition, "пасм"):
		return "☁️"
	case strings.Contains(condition, "облач"):
		return "☁️"
	case strings.Contains(condition, "ясн"):
		return "☀️"
	default:
		return "🌤"
	}
}
