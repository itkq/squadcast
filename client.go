package squadcast

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"go.uber.org/zap"
)

const (
	Endpoint = "https://api.squadcast.com/v3"
)

type Client struct {
	url         *url.URL
	httpClient  *http.Client
	logger      *zap.Logger
	accessToken *AccessToken
}

type AccessToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresAt    int64  `json:"expires_at"`
	IssuedAt     int64  `json:"issued_at"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

func (a *AccessToken) IsExpired() bool {
	return time.Unix(a.ExpiresAt, 0).Before(time.Now())
}

type AccessTokenResponse struct {
	AccessToken AccessToken `json:"data"`
}

type Service struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// More fields ...
}

type ServicesResponse struct {
	Services []*Service `json:"data"`
}

func NewDefaultClient(refreshToken string) (*Client, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	u, _ := url.Parse(Endpoint)

	return &Client{
		url:        u,
		httpClient: &http.Client{},
		logger:     logger,
		accessToken: &AccessToken{
			RefreshToken: refreshToken,
		},
	}, nil
}

func (c *Client) GetAllServices(ctx context.Context) ([]*Service, error) {
	params := &requestParams{
		method: "GET",
		subPath: "/services",
	}

	var servicesResponse ServicesResponse
	if err := c.doAPIRequest(ctx, params, &servicesResponse); err != nil {
		return nil, err
	}

	return servicesResponse.Services, nil
}

func (c *Client) doAPIRequest(ctx context.Context, params *requestParams, out interface{}) error {
	if err := c.authenticate(ctx); err != nil {
		return err
	}

	req, err := c.newRequest(ctx, params)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken.AccessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if err := decodeBodyJSON(resp, &out); err != nil {
		return err
	}

	return nil
}

func (c *Client) authenticate(ctx context.Context) error {
	if c.accessToken == nil || c.accessToken.IsExpired() {
		accessToken, err := c.getAccessToken(ctx)
		if err != nil {
			return nil
		}
		c.accessToken = accessToken
	}

	return nil
}

func (c *Client) getAccessToken(ctx context.Context) (*AccessToken, error) {
	params := &requestParams{
		method: "GET",
		subPath: "/oauth/access-token",
	}
	req, err := c.newRequest(ctx, params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Refresh-Token", c.accessToken.RefreshToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	var accessTokenResponse AccessTokenResponse
	if err := decodeBodyJSON(resp, &accessTokenResponse); err != nil {
		return nil, err
	}

	return &accessTokenResponse.AccessToken, nil
}

func decodeBodyJSON(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

type requestParams struct {
	method string
	subPath string
	body io.Reader
}

func (c *Client) newRequest(ctx context.Context, params *requestParams) (*http.Request, error) {
	u := *c.url
	u.Path = path.Join(c.url.Path, params.subPath)

	req, err := http.NewRequest(params.method, u.String(), params.body)
	if err != nil {
		return nil, err
	}

	return req, nil
}