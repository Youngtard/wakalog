package wakatime

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Youngtard/wakalog/httpclient"
)

func (r *Client) GetSummaries(ctx context.Context, startTime, endTime time.Time) (*Summaries, error) {

	startYear := startTime.Year()
	startMonth := startTime.Month()
	startDay := startTime.Day()

	endYear := endTime.Year()
	endMonth := endTime.Month()
	endDay := endTime.Day()

	u, err := httpclient.ParseURL(baseURL, "/users/current/summaries")

	if err != nil {
		return nil, fmt.Errorf("failed to get summaries fact: %v", err)
	}

	summaries := new(Summaries)

	values := url.Values{}
	values.Add("start", fmt.Sprintf("%d-%d-%d", startYear, startMonth, startDay))
	values.Add("end", fmt.Sprintf("%d-%d-%d", endYear, endMonth, endDay))

	_, err = r.httpclient.Get(ctx, u, values, summaries)

	if err != nil {
		return nil, err
	}

	return summaries, nil

}
