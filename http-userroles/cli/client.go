package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/cloudmachinery/apps/http-userroles/contracts"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	client  *resty.Client
	baseURL string
}

func NewClient(sockAddr string) *Client {
	hc := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", sockAddr)
			},
		},
	}

	rc := resty.NewWithClient(hc)
	rc.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		if resp.StatusCode() >= 400 {
			return fmt.Errorf("http error %d: %s", resp.StatusCode(), resp.Status())
		}
		return nil
	})

	return &Client{
		client:  rc,
		baseURL: "http://unix/api/users",
	}
}

func (c *Client) GetAllUsers() ([]*contracts.User, error) {
	var users []*contracts.User

	_, err := c.client.R().
		SetResult(&users).
		Get(c.baseURL)

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (c *Client) GetUserByEmail(email string) (*contracts.User, error) {
	var user contracts.User

	_, err := c.client.R().
		SetResult(&user).
		Get(fmt.Sprintf("%s/%s", c.baseURL, email))

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *Client) GetUsersByRole(role string) ([]*contracts.User, error) {
	var users []*contracts.User

	_, err := c.client.R().
		SetResult(&users).
		Get(fmt.Sprintf("%s/roles/%s", c.baseURL, role))

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (c *Client) CreateUser(user *contracts.User) error {
	_, err := c.client.R().
		SetBody(user).
		Post(c.baseURL)

	return err
}

func (c *Client) UpdateUser(user *contracts.User) error {
	_, err := c.client.R().
		SetBody(user).
		Put(c.baseURL)

	return err
}

func (c *Client) DeleteUser(email string) error {
	_, err := c.client.R().
		Delete(fmt.Sprintf("%s/%s", c.baseURL, email))

	return err
}
