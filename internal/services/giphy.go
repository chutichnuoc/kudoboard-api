package services

import (
	"encoding/json"
	"fmt"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/utils"
	"net/http"
	"net/url"
)

const (
	// Giphy API base URL
	giphyBaseURL = "https://api.giphy.com/v1/gifs"
)

// GiphyService handles interactions with the Giphy API
type GiphyService struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewGiphyService creates a new Giphy service
func NewGiphyService(cfg *config.Config) *GiphyService {
	return &GiphyService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.HTTPClientTimeout,
		},
	}
}

// Search searches for GIFs based on the provided query parameters
func (s *GiphyService) Search(query string, limit, offset int, rating, lang string) (map[string]interface{}, error) {
	// Build the URL with query parameters
	u, err := url.Parse(fmt.Sprintf("%s/search", giphyBaseURL))
	if err != nil {
		return nil, utils.NewInternalError("Failed to parse Giphy API URL", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("api_key", s.cfg.GiphyApiKey)
	q.Set("q", query)
	q.Set("limit", fmt.Sprintf("%d", limit))
	q.Set("offset", fmt.Sprintf("%d", offset))

	if rating != "" {
		q.Set("rating", rating)
	}
	if lang != "" {
		q.Set("lang", lang)
	}

	u.RawQuery = q.Encode()

	// Create and execute request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, utils.NewInternalError("Failed to create request", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, utils.NewInternalError("Failed to execute request", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, utils.NewInternalError(
			fmt.Sprintf("Giphy API returned non-OK status: %d", resp.StatusCode),
			fmt.Errorf("status code: %d", resp.StatusCode),
		)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, utils.NewInternalError("Failed to parse Giphy response", err)
	}

	return result, nil
}

// Trending gets trending GIFs
func (s *GiphyService) Trending(limit, offset int, rating string) (map[string]interface{}, error) {
	// Build the URL with query parameters
	u, err := url.Parse(fmt.Sprintf("%s/trending", giphyBaseURL))
	if err != nil {
		return nil, utils.NewInternalError("Failed to parse Giphy API URL", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("api_key", s.cfg.GiphyApiKey)
	q.Set("limit", fmt.Sprintf("%d", limit))
	q.Set("offset", fmt.Sprintf("%d", offset))

	if rating != "" {
		q.Set("rating", rating)
	}

	u.RawQuery = q.Encode()

	// Create and execute request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, utils.NewInternalError("Failed to create request", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, utils.NewInternalError("Failed to execute request", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, utils.NewInternalError(
			fmt.Sprintf("Giphy API returned non-OK status: %d", resp.StatusCode),
			fmt.Errorf("status code: %d", resp.StatusCode),
		)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, utils.NewInternalError("Failed to parse Giphy response", err)
	}

	return result, nil
}

// GetById gets a specific GIF by ID
func (s *GiphyService) GetById(gifId string) (map[string]interface{}, error) {
	// Build the URL
	u, err := url.Parse(fmt.Sprintf("%s/%s", giphyBaseURL, gifId))
	if err != nil {
		return nil, utils.NewInternalError("Failed to parse Giphy API URL", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("api_key", s.cfg.GiphyApiKey)
	u.RawQuery = q.Encode()

	// Create and execute request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, utils.NewInternalError("Failed to create request", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, utils.NewInternalError("Failed to execute request", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == http.StatusNotFound {
		return nil, utils.NewNotFoundError("GIF not found")
	} else if resp.StatusCode != http.StatusOK {
		return nil, utils.NewInternalError(
			fmt.Sprintf("Giphy API returned non-OK status: %d", resp.StatusCode),
			fmt.Errorf("status code: %d", resp.StatusCode),
		)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, utils.NewInternalError("Failed to parse Giphy response", err)
	}

	return result, nil
}

// Random gets a random GIF
func (s *GiphyService) Random(tag string, rating string) (map[string]interface{}, error) {
	// Build the URL with query parameters
	u, err := url.Parse(fmt.Sprintf("%s/random", giphyBaseURL))
	if err != nil {
		return nil, utils.NewInternalError("Failed to parse Giphy API URL", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("api_key", s.cfg.GiphyApiKey)

	if tag != "" {
		q.Set("tag", tag)
	}

	if rating != "" {
		q.Set("rating", rating)
	}

	u.RawQuery = q.Encode()

	// Create and execute request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, utils.NewInternalError("Failed to create request", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, utils.NewInternalError("Failed to execute request", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, utils.NewInternalError(
			fmt.Sprintf("Giphy API returned non-OK status: %d", resp.StatusCode),
			fmt.Errorf("status code: %d", resp.StatusCode),
		)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, utils.NewInternalError("Failed to parse Giphy response", err)
	}

	return result, nil
}
