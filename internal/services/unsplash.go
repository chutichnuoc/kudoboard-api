package services

import (
	"encoding/json"
	"fmt"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/utils"
	"net/http"
	"net/url"
	"strconv"
)

const (
	// Unsplash API base URL
	unsplashBaseURL = "https://api.unsplash.com"
)

// UnsplashService handles interactions with the Unsplash API
type UnsplashService struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewUnsplashService creates a new Unsplash service
func NewUnsplashService(cfg *config.Config) *UnsplashService {
	return &UnsplashService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.HTTPClientTimeout,
		},
	}
}

// Search searches for photos based on the provided query parameters
func (s *UnsplashService) Search(query string, page, perPage int, orderBy string) (map[string]interface{}, error) {
	// Build the URL with query parameters
	u, err := url.Parse(fmt.Sprintf("%s/search/photos", unsplashBaseURL))
	if err != nil {
		return nil, utils.NewInternalError("Failed to parse Unsplash API URL", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("query", query)
	q.Set("page", strconv.Itoa(page))
	q.Set("per_page", strconv.Itoa(perPage))

	if orderBy != "" {
		q.Set("order_by", orderBy)
	}

	u.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, utils.NewInternalError("Failed to create request", err)
	}

	// Add required headers
	req.Header.Add("Authorization", fmt.Sprintf("Client-ID %s", s.cfg.UnsplashAccessKey))
	req.Header.Add("Accept-Version", "v1")

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, utils.NewInternalError("Failed to execute request", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, utils.NewUnauthorizedError("Invalid Unsplash API credentials")
	} else if resp.StatusCode != http.StatusOK {
		return nil, utils.NewInternalError(
			fmt.Sprintf("Unsplash API returned non-OK status: %d", resp.StatusCode),
			fmt.Errorf("status code: %d", resp.StatusCode),
		)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, utils.NewInternalError("Failed to parse Unsplash response", err)
	}

	return result, nil
}

// Random gets random photos, optionally filtered by topics or collections
func (s *UnsplashService) Random(count int, query, topics, username, collections string, featured bool) (map[string]interface{}, error) {
	// Build the URL with query parameters
	u, err := url.Parse(fmt.Sprintf("%s/photos/random", unsplashBaseURL))
	if err != nil {
		return nil, utils.NewInternalError("Failed to parse Unsplash API URL", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("count", strconv.Itoa(count))

	if query != "" {
		q.Set("query", query)
	}

	if topics != "" {
		q.Set("topics", topics)
	}

	if username != "" {
		q.Set("username", username)
	}

	if collections != "" {
		q.Set("collections", collections)
	}

	if featured {
		q.Set("featured", "true")
	}

	u.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, utils.NewInternalError("Failed to create request", err)
	}

	// Add required headers
	req.Header.Add("Authorization", fmt.Sprintf("Client-ID %s", s.cfg.UnsplashAccessKey))
	req.Header.Add("Accept-Version", "v1")

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, utils.NewInternalError("Failed to execute request", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, utils.NewUnauthorizedError("Invalid Unsplash API credentials")
	} else if resp.StatusCode != http.StatusOK {
		return nil, utils.NewInternalError(
			fmt.Sprintf("Unsplash API returned non-OK status: %d", resp.StatusCode),
			fmt.Errorf("status code: %d", resp.StatusCode),
		)
	}

	// Parse response - can be an array or an object
	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, utils.NewInternalError("Failed to parse Unsplash response", err)
	}

	// Wrap array results in an object
	var finalResult map[string]interface{}
	switch v := result.(type) {
	case []interface{}:
		finalResult = map[string]interface{}{
			"results": v,
		}
	case map[string]interface{}:
		finalResult = v
	default:
		return nil, utils.NewInternalError("Unexpected response format from Unsplash", nil)
	}

	return finalResult, nil
}

// GetById gets a specific photo by ID
func (s *UnsplashService) GetById(photoID string) (map[string]interface{}, error) {
	// Build the URL
	u, err := url.Parse(fmt.Sprintf("%s/photos/%s", unsplashBaseURL, photoID))
	if err != nil {
		return nil, utils.NewInternalError("Failed to parse Unsplash API URL", err)
	}

	// Create request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, utils.NewInternalError("Failed to create request", err)
	}

	// Add required headers
	req.Header.Add("Authorization", fmt.Sprintf("Client-ID %s", s.cfg.UnsplashAccessKey))
	req.Header.Add("Accept-Version", "v1")

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, utils.NewInternalError("Failed to execute request", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, utils.NewUnauthorizedError("Invalid Unsplash API credentials")
	} else if resp.StatusCode == http.StatusNotFound {
		return nil, utils.NewNotFoundError("Photo not found")
	} else if resp.StatusCode != http.StatusOK {
		return nil, utils.NewInternalError(
			fmt.Sprintf("Unsplash API returned non-OK status: %d", resp.StatusCode),
			fmt.Errorf("status code: %d", resp.StatusCode),
		)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, utils.NewInternalError("Failed to parse Unsplash response", err)
	}

	return result, nil
}
