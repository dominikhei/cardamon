package grafana

import (
	"encoding/json"
	"io"
	"net/http"
)

type Client struct {
	URL    string
	APIKey string
	Client *http.Client
}

type DashboardMetadata struct {
	UID   string `json:"uid"`
	Title string `json:"title"`
}

func NewClient(url, apiKey string) *Client {
	return &Client{
		URL:    url,
		APIKey: apiKey,
		Client: &http.Client{},
	}
}

func (c *Client) newRequest(method, path string) (*http.Request, error) {
	req, err := http.NewRequest(method, c.URL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// SearchDashboards finds all dashboard UIDs
func (c *Client) SearchDashboards() ([]DashboardMetadata, error) {
	req, err := c.newRequest("GET", "/api/search?type=dash-db")
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []DashboardMetadata
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetDashboardModel fetches the raw JSON of a dashboard
func (c *Client) GetDashboardModel(uid string) ([]byte, error) {
	req, err := c.newRequest("GET", "/api/dashboards/uid/"+uid)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}