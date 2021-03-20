package squadcast

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

const (
	WebhookEndpoint = "https://api.squadcast.com/v2/incidents/api"
)

type WebhookClient struct {
	apiKey     string
	httpClient *http.Client
	url        *url.URL
}

func NewDefaultWebhookClient(apiKey string) *WebhookClient {
	u, _ := url.Parse(WebhookEndpoint)

	return &WebhookClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		url:        u,
	}
}

func NewTestWebhookClient(endpoint string, apiKey string) *WebhookClient {
	u, _ := url.Parse(endpoint)

	return &WebhookClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		url:        u,
	}
}

type PostIncidentRequest struct {
	Message     string
	Description string
	Status      string
	// more fields ...
}

func (c *WebhookClient) PostIncident(ctx context.Context, request *PostIncidentRequest) error {
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}

	params := &requestParams{
		method:  "POST",
		subPath: c.apiKey,
		body:    bytes.NewBuffer(b),
	}

	req, err := c.newRequest(ctx, params)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *WebhookClient) newRequest(ctx context.Context, params *requestParams) (*http.Request, error) {
	u := *c.url
	u.Path = path.Join(c.url.Path, params.subPath)

	req, err := http.NewRequest(params.method, u.String(), params.body)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for k, v := range params.queries {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	return req, nil
}
