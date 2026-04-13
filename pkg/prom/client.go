// Copyright 2026 dominikhei
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
