package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// URL is the default URL for the host of Enterprise.
// This variable can be set globally or on a per Client
// instance.
var URL = "https://enterprise.influxdata.com"

// Client handles all of the heavy lifting of talking
// to Enterprise for you.
type Client struct {
	URL   string // Defaults to `client.URL`
	Token string // OPTIONAL: The token of the customer making the request
}

// New returns a configured `Client`. The `token`
// is optional, but if you have it, you should pass
// it in.
func New(token string) *Client {
	return &Client{
		URL:   URL,
		Token: token,
	}
}

type saveable interface {
	path() string
}

func (c *Client) Save(s saveable) (*http.Response, error) {
	u := fmt.Sprintf("%s/api/v1%s", c.URL, s.path())
	fmt.Printf("u: %s\n", u)

	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("X-Authorization", c.Token)
	}

	cl := http.Client{}
	res, err := cl.Do(req)
	if err != nil {
		return res, err
	}

	code := res.StatusCode
	switch code {
	case 401, 404, 500:
		se := SimpleError{}
		err = json.NewDecoder(res.Body).Decode(&se)
		if err != nil {
			return res, err
		}
		return res, se
	case 422:
		ve := ValidationErrors{}
		err = json.NewDecoder(res.Body).Decode(&ve)
		if err != nil {
			return res, err
		}
		return res, ve
	}

	return res, err
}