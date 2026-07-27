package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	appErr "atlas/internal/errors"
)

type Termii struct {
	APIKey  string
	BaseURL string
	From    string
	Client  *http.Client
}

func NewTermii(apiKey, baseURL, from string) (*Termii, error) {
	if apiKey == "" || baseURL == "" || from == "" {
		return nil, appErr.ErrInvalidTermiiConfig
	}
	return &Termii{
		APIKey:  apiKey,
		BaseURL: baseURL,
		From:    from,
		Client:  &http.Client{Timeout: time.Second * 15},
	}, nil
}

func (t *Termii) Send(ctx context.Context, to, message string, channel Channel) error {
	if to == "" || message == "" {
		return appErr.ErrInvalidTermiiConfig
	}

	url := strings.TrimSpace(t.BaseURL) + "/api/sms/send"

	payload := TermiiRequestBody{
		APIKey:  t.APIKey,
		To:      to,
		From:    t.From,
		SMS:     message,
		Channel: channel,
		Type:    "plain",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return &appErr.TermiiError{Status: http.StatusInternalServerError, Message: "Failed to marshal payload"}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return &appErr.TermiiError{Status: http.StatusInternalServerError, Message: "Failed to create request"}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := t.Client.Do(req)
	if err != nil {
		return &appErr.TermiiError{Status: http.StatusInternalServerError, Message: "Failed to send request"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return appErr.ErrTermiiSendFailed
	}

	return nil
}
