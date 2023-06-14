package music

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Result represents a single search result item.
type Result struct {
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	ReleaseYear string `json:"release_year"`
	Source      string `json:"source"`
}

// Search performs a music search query on Deezer API and iTunes API and returns the combined search results.
func Search(query string) ([]Result, error) {
	deezerResults, err := searchDeezer(query)
	if err != nil {
		return nil, err
	}

	itunesResults, err := searchiTunes(query)
	if err != nil {
		return nil, err
	}

	// Combine the results from both APIs
	results := append(deezerResults, itunesResults...)
	return results, nil
}

// searchDeezer performs a music search query on the Deezer API and returns the search results.
func searchDeezer(query string) ([]Result, error) {
	// Set up the API endpoint URL
	baseURL := "https://api.deezer.com/search"
	searchURL := fmt.Sprintf("%s?q=%s", baseURL, query)

	// Send the HTTP GET request
	response, err := http.Get(searchURL)
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

// searchiTunes performs a music search query on the iTunes API and returns the search results.
func searchiTunes(query string) ([]Result, error) {
	// Set up the API endpoint URL
	baseURL := "https://itunes.apple.com/search"
	searchURL := fmt.Sprintf("%s?term=%s&entity=song", baseURL, query)

	// Send the HTTP GET request
	response, err := http.Get(searchURL)
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
	var data struct {
		Results []struct {
			TrackName      string `json:"trackName"`
			ArtistName     string `json:"artistName"`
			CollectionName string `json:"collectionName"`
			ReleaseDate    string `json:"releaseDate"`
		} `json:"results"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	// Extract the search results
	results := make([]Result, 0)
	for _, track := range data.Results {
		result := Result{
			Title:       track.TrackName,
			Artist:      track.ArtistName,
			Album:       track.CollectionName,
			ReleaseYear: track.ReleaseDate[:4],
			Source:      "iTunes",
		}
		results = append(results, result)
	}

	return results, nil
}
