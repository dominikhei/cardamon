package prom

import (
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// Contains the Prometheus client.
type Client struct {
	api v1.API
}

// Initializes a Prometheus client using the Prometheus go package.
func NewClient(address string) (*Client, error) {
	c, err := api.NewClient(api.Config{Address: address})
	if err != nil {
		return nil, err
	}
	return &Client{api: v1.NewAPI(c)}, nil
}