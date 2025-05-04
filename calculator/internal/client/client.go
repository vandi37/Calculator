package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/vandi37/Calculator/internal/models"
	"github.com/vandi37/ferror"
)

type Client struct {
	Address    string
	httpClient *http.Client
}

func New(address string) *Client {
	return &Client{
		Address:    address,
		httpClient: &http.Client{},
	}
}

func (c *Client) Request(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c *Client) Read(resp *http.Response, v any) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

func (c *Client) ReadError(resp *http.Response) error {
	v := new(models.ErrorResponse)
	if err := c.Read(resp, v); err != nil {
		return err
	}
	return ferror.Simple("Server", v.Error)
}

func (c *Client) Login(ctx context.Context, username, password string) (*Profile, error) {
	b, err := json.Marshal(models.UserRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Address+"/login", strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	resp, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.ReadError(resp)
	}
	token := new(models.TokenResponse)
	if err := c.Read(resp, token); err != nil {
		return nil, err
	}
	return &Profile{
		token:  token.Token,
		Client: c,
	}, nil
}

func (c *Client) Register(ctx context.Context, username, password string) (*Profile, error) {
	b, err := json.Marshal(models.UserRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Address+"/register", strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	resp, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, c.ReadError(resp)
	}
	// We don't need the response
	resp.Body.Close()
	return c.Login(ctx, username, password)
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.Address+"/ping", nil)
	if err != nil {
		return err
	}
	resp, err := c.Request(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return c.ReadError(resp)
	}
	return nil
}
