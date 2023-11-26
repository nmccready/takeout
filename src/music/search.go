package music

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/nmccready/oauth2"
	_oauth2 "github.com/nmccready/takeout/src/oauth2"
)

/*
	searchDeezer performs a music search query on the Deezer API and returns the search results.

https://api.deezer.com/search?q=eminem
*/
func searchDeezer(opts SearchOpts) ([]Result, error) {
	// Create an HTTP client with OAuth2 authentication
	token, err := _oauth2.GetDeezerToken()
	if err != nil {
		return nil, err
	}

	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	// Set up the API endpoint URL
	baseURL := "https://api.deezer.com/search"
	searchURL := fmt.Sprintf("%s?%s", baseURL, opts.deezerEncode())

	// Send the HTTP GET request
	response, err := client.Get(searchURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	// Extract the search results
	results := make([]Result, 0)
	if tracks, ok := data["data"].([]interface{}); ok {
		for _, track := range tracks {
			trackData := track.(map[string]interface{})
			result := Result{
				Title:       trackData["title"].(string),
				Artist:      trackData["artist"].(map[string]interface{})["name"].(string),
				Album:       trackData["album"].(map[string]interface{})["title"].(string),
				ReleaseYear: trackData["album"].(map[string]interface{})["release_date"].(string)[:4],
				Source:      "Deezer",
			}
			results = append(results, result)
		}
	}

	return results, nil
}
