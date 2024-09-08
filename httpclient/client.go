package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	client *http.Client
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func NewClient(client *http.Client) *Client {

	if client == nil {
		client = http.DefaultClient
	}

	return &Client{
		client: client,
	}

}

func (c *Client) WithAuthToken(token string) *Client {

	c2 := c.copy()
	transport := c2.client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	c2.client.Transport = roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			req = req.Clone(req.Context())
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			return transport.RoundTrip(req)
		},
	)
	return c2
}

func (c *Client) copy() *Client {

	clone := Client{
		client: &http.Client{},
	}

	if c.client != nil {
		clone.client.Transport = c.client.Transport
		clone.client.CheckRedirect = c.client.CheckRedirect
		clone.client.Jar = c.client.Jar
		clone.client.Timeout = c.client.Timeout
	}

	return &clone
}

type ServerError struct {
	Body       io.ReadCloser
	StatusCode int
}

func (se *ServerError) Error() string {

	return "unexpected status code"

}

func (c *Client) Get(ctx context.Context, urlPath string, params url.Values, v interface{}) (*http.Response, error) {
	url, err := url.Parse(urlPath)

	if err != nil {
		return nil, fmt.Errorf("error parsing url path")
	}
	url.Query()

	url.RawQuery = params.Encode()

	req, err := c.createRequest(http.MethodGet, url.String(), nil)

	if err != nil {
		return nil, err
	}

	resp, err := c.processRequest(ctx, req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = decodeResponse(resp.Body, v)

	if err != nil {
		return nil, err
	}

	return resp, nil

}

func (c *Client) createRequest(method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, fmt.Errorf("error creating request")
	}

	return req, nil

}

func (c *Client) processRequest(ctx context.Context, req *http.Request) (*http.Response, error) {

	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)

	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return nil, fmt.Errorf("error making request: %w", err)
		}

	}

	success := resp.StatusCode >= 200 && resp.StatusCode < 400

	if !success {
		return nil, &ServerError{Body: resp.Body, StatusCode: resp.StatusCode}
	}

	return resp, nil
}
