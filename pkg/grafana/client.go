package grafana

import (
	"encoding/json"
	"io"
	"net/http"
)

// Client contains all necessary params to connect to Grafana.
type Client struct {
	URL        string
	PathPrefix string
	Token      string
	Username   string
	Password   string
	Client     *http.Client
}

// DashboardMetadata holds UIDs and Titels of Dashboards.
type DashboardMetadata struct {
	UID   string `json:"uid"`
	Title string `json:"title"`
}

func NewClient(url, pathPrefix string, token string, username string, password string) *Client {
	return &Client{
		URL:        url,
		PathPrefix: pathPrefix,
		Token:      token,
		Username:   username,
		Password:   password,
		Client:     &http.Client{},
	}
}

// newRequest is used to make a request to the Grafana instance, based on the params in the client struct.
func (c *Client) newRequest(method, path string) (*http.Request, error) {
	req, err := http.NewRequest(method, c.URL+c.PathPrefix+path, nil)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	} else if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// SearchDashboards finds all dashboard UIDs.
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

// GetDashboardModel fetches the raw JSON of a dashboard based on a UID.
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
