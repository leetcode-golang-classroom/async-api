package report

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const BaseURL = "https://botw-compendium.herokuapp.com/api/v3/compendium"

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type LozClient struct {
	baseURL    string
	httpClient HTTPClient
}

func NewClient(httpClient HTTPClient) *LozClient {
	return &LozClient{
		baseURL:    BaseURL,
		httpClient: httpClient,
	}
}

type Monster struct {
	Name            string   `json:"name"`
	ID              int32    `json:"id"`
	Category        string   `json:"category"`
	Description     string   `json:"description"`
	Image           string   `json:"image"`
	CommonLocations []string `json:"common_locations"`
	Drops           []string `json:"drops"`
	Dlc             bool     `json:"dlc"`
}
type GetMonstersResponse struct {
	Data []Monster `json:"data"`
}

func (c *LozClient) GetMonsters() (*GetMonstersResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/category/monsters", c.baseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create monsters request: %w", err)
	}

	reqURL := req.URL
	queryParams := req.URL.Query()
	queryParams.Set("game", "totk")
	reqURL.RawQuery = queryParams.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit monsters http request: %w", err)
	}

	var response GetMonstersResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal monsters http response: %w", err)
	}
	return &response, nil
}
