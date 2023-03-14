package recaptcha

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/altipla-consulting/errors"
)

type apiResponse struct {
	Success bool `json:"success"`
}

type Client struct {
	privateKey string
}

func NewClient(privateKey string) *Client {
	return &Client{privateKey}
}

func (client *Client) Validate(ctx context.Context, caption string) (bool, error) {
	q := make(url.Values)
	q.Set("secret", client.privateKey)
	q.Set("response", caption)
	u := &url.URL{
		Scheme:   "https",
		Host:     "www.google.com",
		Path:     "/recaptcha/api/siteverify",
		RawQuery: q.Encode(),
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return false, errors.Trace(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, errors.Errorf("bad recaptcha http status: %s", resp.Status)
	}

	reply := new(apiResponse)
	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return false, errors.Trace(err)
	}

	return reply.Success, nil
}
