package wakatime

import "time"

type Summaries struct {
	Data            []SummariesData `json:"data"`
	Start           time.Time       `json:"start"`
	End             time.Time       `json:"end"`
	CumulativeTotal CumulativeTotal `json:"cumulative_total"`
	DailyAverage    DailyAverage    `json:"daily_average"`
}

type SummariesData struct {
	GrandTotal GrandTotal `json:"grand_total"`
	Projects   []Project  `json:"projects"`
}

type GrandTotal struct {
	Hours        int     `json:"hours"`
	Minutes      int     `json:"minutes"`
	TotalSeconds float64 `json:"total_seconds"`
	Digital      string  `json:"digital"`
	Decimal      string  `json:"decimal"`
	Text         string  `json:"text"`
}

type Project struct {
	Name          string      `json:"name"`
	TotalSeconds  float64     `json:"total_seconds"`
	Digital       string      `json:"digital"`
	Decimal       string      `json:"decimal"`
	Text          string      `json:"text"`
	Hours         int64       `json:"hours"`
	Minutes       int64       `json:"minutes"`
	Seconds       int64       `json:"seconds"`
	Percent       float64     `json:"percent"`
	MachineNameID *string     `json:"machine_name_id,omitempty"`
	Color         interface{} `json:"color"`
}

type CumulativeTotal struct {
	Seconds float64 `json:"seconds"`
	Text    string  `json:"text"`
	Digital string  `json:"digital"`
	Decimal string  `json:"decimal"`
}
type DailyAverage struct {
	Holidays                      int    `json:"holidays"`
	DaysMinusHolidays             int    `json:"days_minus_holidays"`
	DaysIncludingHolidays         int    `json:"days_including_holidays"`
	Seconds                       int    `json:"seconds"`
	SecondsIncludingOtherLanguage int    `json:"seconds_including_other_language"`
	Text                          string `json:"text"`
	TextIncludingOtherLanguage    string `json:"text_including_other_language"`
}
