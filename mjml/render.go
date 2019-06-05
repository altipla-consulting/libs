package mjml

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"libs.altipla.consulting/errors"
)

type renderRequest struct {
	Content string `json:"template"`
}

type renderReply struct {
	Content string `json:"content"`
}

func Render(ctx context.Context, content string) (string, error) {
	data := &renderRequest{
		Content: content,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		return "", errors.Trace(err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://europe-west1-precise-truck-89123.cloudfunctions.net/render-mjml", &buf)
	if err != nil {
		return "", errors.Trace(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("bad render-mjml status: %s", resp.Status)
	}

	reply := new(renderReply)
	if err = json.NewDecoder(resp.Body).Decode(reply); err != nil {
		return "", errors.Trace(err)
	}

	return reply.Content, nil
}
