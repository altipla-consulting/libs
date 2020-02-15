package mjml

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"libs.altipla.consulting/errors"
)

type renderRequest struct {
	Content string `json:"template"`
}

type renderReply struct {
	Content string `json:"content"`
}

func Render(ctx context.Context, content string) (string, error) {
	var lastErr string
	for ctx.Err() == nil {
		result, err := renderShort(ctx, content)
		if err != nil {
			lastErr = err.Error()
			if ctx.Err() != nil {
				return "", errors.Trace(ctx.Err())
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}

		return result, nil
	}

	return "", errors.Wrapf(ctx.Err(), "last error: %v", lastErr)
}

func renderShort(ctx context.Context, content string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

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
