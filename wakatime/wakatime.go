package wakatime

import "github.com/Youngtard/wakalog/httpclient"

const (
	apiVersion = "/api/v1"
	baseURL    = "https://api.wakatime.com" + apiVersion
)

type Client struct {
	httpclient *httpclient.Client
}

func NewClient(hClient *httpclient.Client) *Client {

	c := &Client{
		httpclient: hClient,
	}
	return c

}
