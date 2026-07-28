package weather

type CurrentResponse struct {
	Location Location `json:"location"`
	Current  Current  `json:"current"`
}

type Location struct {
	Name    string `json:"name"`
	Country string `json:"country"`
	Local   string `json:"localtime"`
}

type Current struct {
	TempC      float64   `json:"temp_c"`
	FeelsLikeC float64   `json:"feelslike_c"`
	WindKph    float64   `json:"wind_kph"`
	Humidity   int       `json:"humidity"`
	Cloud      int       `json:"cloud"`
	Condition  Condition `json:"condition"`
}

type Condition struct {
	Text string `json:"text"`
	Icon string `json:"icon"`
}

type ErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type ForecastResponse struct {
	Location Location `json:"location"`
	Current  Current  `json:"current"`
	Forecast struct {
		Forecastday []ForecastDay `json:"forecastday"`
	} `json:"forecast"`
}

type ForecastDay struct {
	Date string `json:"date"`

	Day struct {
		MinTemp   float64   `json:"mintemp_c"`
		MaxTemp   float64   `json:"maxtemp_c"`
		Condition Condition `json:"condition"`
	} `json:"day"`

	Hour []Hour `json:"hour"`
}

type Hour struct {
	Time      string    `json:"time"`
	Temp      float64   `json:"temp_c"`
	Condition Condition `json:"condition"`
}
