package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)


func HttpGet(uri string, headers map[string]string, responseBody interface{}) error {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil { return fmt.Errorf("failed to create GET(%s) request: %w", uri, err) }

	// Add custom headers
	for key, value := range headers { req.Header.Set(key, value) }

	// Make the request
	res, err := http.DefaultClient.Do(req)
	if err != nil { return fmt.Errorf("failed to make GET(%s) request: %w", uri, err) }
	defer res.Body.Close()

	// Read the response as a byte slice
    if res.StatusCode < 200 || res.StatusCode >= 300 {
        bodyBytes, _ := io.ReadAll(res.Body)
        return fmt.Errorf("Get(%s) request failed with status %d: %s", uri, res.StatusCode, string(bodyBytes))
    }

	// Convert byte slice to string and return
	if err := json.NewDecoder(res.Body).Decode(responseBody); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return nil
}

func HttpPost(uri string, headers map[string]string, requestBody interface{}, responseBody interface{}) error {
	// Convert request body to JSON
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil { return fmt.Errorf("failed to marshal request body: %w", err) }

	// Create request
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(requestBodyBytes))
	if err != nil { return fmt.Errorf("failed to create POST(%s) request: %w", uri, err) }

	// Add custom headers
	for key, value := range headers { req.Header.Set(key, value) }

	// Make the request
	res, err := http.DefaultClient.Do(req)
	if err != nil { return fmt.Errorf("failed to make POST(%s) request: %w", uri, err) }
	defer res.Body.Close()

	// Read the response as a byte slice
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("POST(%s) request failed with status %d: %s", uri, res.StatusCode, string(bodyBytes))
	}

	// Convert byte slice to string and return
	if err := json.NewDecoder(res.Body).Decode(responseBody); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return nil
}