package httpclient

import (
	"encoding/json"
	"io"
	"net/url"
)

func ParseURL(baseURL, urlPath string) (string, error) {
	parsedURL, err := url.JoinPath(baseURL, urlPath)

	if err != nil {
		return "", err
	}

	return parsedURL, nil
}

// https://github.com/google/go-github/blob/master/github/github.go
func decodeResponse(res io.Reader, v interface{}) error {

	var err error

	switch v := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(v, res)
	default:
		decErr := json.NewDecoder(res).Decode(v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}

	return err

}
