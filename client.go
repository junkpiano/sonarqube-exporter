package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	endpoint           *url.URL
	username, password string
}

func NewClient(endpoint, username, password string) (*Client, error) {
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}

	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	c := &Client{endpoint: url, username: username, password: password}
	return c, nil
}

func (c *Client) newRequest(method, path string) ([]byte, error) {
	c.endpoint.Path = path

	req, err := http.NewRequest(method, c.endpoint.String(), nil)

	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(c.username, c.password)

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type ActivityStatus struct {
	Pending    int `json: "pending"`
	Failing    int `json: "failing`
	InProgress int `json: "inProgress"`
}

func (c *Client) ActivityStatus() (*ActivityStatus, error) {
	b, err := c.newRequest(http.MethodGet, "api/ce/activity_status")

	as := ActivityStatus{}

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &as)

	if err != nil {
		return nil, err
	}

	return &as, nil
}
