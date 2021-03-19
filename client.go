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
	var servicesResponse ServicesResponse
	if err := c.doAPIRequest(ctx, "GET", "/services", nil, &servicesResponse); err != nil {
		return nil, err
	}

	return servicesResponse.Services, nil
}

func (c *Client) doAPIRequest(ctx context.Context, method, spath string, body io.Reader, iface interface{}) error {
	if err := c.authenticate(ctx); err != nil {
		return err
	}

	req, err := c.newRequest(ctx, method, spath, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken.AccessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if err := decodeBodyJSON(resp, &iface); err != nil {
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
	req, err := c.newRequest(ctx, "GET", "/oauth/access-token", nil)
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

func (c *Client) newRequest(ctx context.Context, method, spath string, body io.Reader) (*http.Request, error) {
	u := *c.url
	u.Path = path.Join(c.url.Path, spath)

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	return req, nil
}
