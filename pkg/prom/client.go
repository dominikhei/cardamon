package prom

import (
	"net/http"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type Client struct {
	api v1.API
}

type roundTripper struct {
	token    string
	username string
	password string
	base     http.RoundTripper
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	if rt.token != "" {
		r.Header.Set("Authorization", "Bearer "+rt.token)
	} else if rt.username != "" {
		r.SetBasicAuth(rt.username, rt.password)
	}
	return rt.base.RoundTrip(r)
}

func NewClient(address, token, username, password string) (*Client, error) {
	c, err := api.NewClient(api.Config{
		Address: address,
		RoundTripper: &roundTripper{
			token:    token,
			username: username,
			password: password,
			base:     api.DefaultRoundTripper,
		},
	})
	if err != nil {
		return nil, err
	}
	return &Client{api: v1.NewAPI(c)}, nil
}
