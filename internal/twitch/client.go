// Package twitch provides Twitch API integration for stream status monitoring.
package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents a Twitch API client
type Client struct {
	clientID     string
	clientSecret string
	accessToken  string
	httpClient   *http.Client
}

// Config holds Twitch API configuration
type Config struct {
	ClientID     string
	ClientSecret string
}

// StreamData represents Twitch stream information
type StreamData struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserLogin    string    `json:"user_login"`
	UserName     string    `json:"user_name"`
	GameID       string    `json:"game_id"`
	GameName     string    `json:"game_name"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailURL string    `json:"thumbnail_url"`
	TagIDs       []string  `json:"tag_ids"`
	IsMature     bool      `json:"is_mature"`
}

// StreamsResponse represents the API response for streams
type StreamsResponse struct {
	Data       []StreamData `json:"data"`
	Pagination struct {
		Cursor string `json:"cursor"`
	} `json:"pagination"`
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// NewClient creates a new Twitch API client
func NewClient(config Config) (*Client, error) {
	client := &Client{
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Get OAuth token
	if err := client.authenticate(); err != nil {
		return nil, fmt.Errorf("failed to authenticate with Twitch API: %w", err)
	}

	return client, nil
}

// authenticate gets an OAuth token using client credentials flow
func (c *Client) authenticate() error {
	tokenURL := "https://id.twitch.tv/oauth2/token"

	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	return nil
}

// GetStreamByUsername checks if a specific user is currently streaming
func (c *Client) GetStreamByUsername(username string) (*StreamData, error) {
	apiURL := fmt.Sprintf("https://api.twitch.tv/helix/streams?user_login=%s", url.QueryEscape(username))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create streams request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Client-Id", c.clientID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send streams request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Token might be expired, try to re-authenticate
		if authErr := c.authenticate(); authErr != nil {
			return nil, fmt.Errorf("failed to re-authenticate: %w", authErr)
		}
		// Retry the request with new token
		return c.GetStreamByUsername(username)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("streams request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var streamsResp StreamsResponse
	if err := json.NewDecoder(resp.Body).Decode(&streamsResp); err != nil {
		return nil, fmt.Errorf("failed to decode streams response: %w", err)
	}

	// If no streams found, user is offline
	if len(streamsResp.Data) == 0 {
		return nil, nil
	}

	return &streamsResp.Data[0], nil
}

// GetMultipleStreams checks multiple usernames at once (more efficient)
func (c *Client) GetMultipleStreams(usernames []string) ([]StreamData, error) {
	if len(usernames) == 0 {
		return []StreamData{}, nil
	}

	// Twitch API allows up to 100 usernames per request
	if len(usernames) > 100 {
		usernames = usernames[:100]
	}

	// Build query parameters
	params := url.Values{}
	for _, username := range usernames {
		params.Add("user_login", username)
	}

	apiURL := "https://api.twitch.tv/helix/streams?" + params.Encode()

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create streams request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Client-Id", c.clientID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send streams request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Token might be expired, try to re-authenticate
		if authErr := c.authenticate(); authErr != nil {
			return nil, fmt.Errorf("failed to re-authenticate: %w", authErr)
		}
		// Retry the request with new token
		return c.GetMultipleStreams(usernames)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("streams request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var streamsResp StreamsResponse
	if err := json.NewDecoder(resp.Body).Decode(&streamsResp); err != nil {
		return nil, fmt.Errorf("failed to decode streams response: %w", err)
	}

	return streamsResp.Data, nil
}

// IsUserLive is a convenience method to check if a single user is live
func (c *Client) IsUserLive(username string) (bool, *StreamData, error) {
	stream, err := c.GetStreamByUsername(username)
	if err != nil {
		return false, nil, err
	}

	if stream == nil {
		return false, nil, nil
	}

	return true, stream, nil
}
