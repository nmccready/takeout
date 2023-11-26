package music

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/nmccready/takeout/src/internal/logger"
)

var debug = logger.Spawn("music")

// Result represents a single search result item.
type Result struct {
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	ReleaseYear string `json:"release_year"`
	Source      string `json:"source"`
}

// Search performs a music search query on Deezer API and iTunes API and returns the combined search results.
func Search(opts SearchOpts) ([]Result, error) {
	deezerResults, err := searchDeezer(opts)
	if err != nil {
		debug.Error("searchDeezer err: %s", err.Error())
		return nil, err
	}

	// itunesResults, err := searchAppleMusic(query)
	// if err != nil {
	// 	return nil, err
	// }

	// Combine the results from both APIs
	results := append(deezerResults) //, itunesResults...)
	return results, nil
}

type SearchOpts struct {
	Title  string
	Album  string
	Artist string
	Year   string
}

func deezerKeyValue(key, value string) string {
	return fmt.Sprintf("%s:\"%s\"", key, value)
}

// deezerEncode encodes the search options into a Deezer API query string.
// q=track:"eminem" album:"curtain call" artist:"eminem"
func (opts SearchOpts) deezerEncode() string {
	query := url.Values{}
	keyValues := map[string]string{}
	if opts.Album != "" {
		keyValues["album"] = opts.Album
	}
	if opts.Artist != "" {
		keyValues["artist"] = opts.Artist
	}
	if opts.Title != "" {
		keyValues["tack"] = opts.Title
	}

	queryStr := ""
	for key, value := range keyValues {
		if queryStr != "" {
			queryStr += " "
		}
		queryStr += deezerKeyValue(key, value)
	}

	query.Set("q", queryStr)

	return query.Encode()
}

func (opts SearchOpts) appleMusicEncode() string {
	query := url.Values{}
	if opts.Title != "" {
		query.Set("term", opts.Title)
	}
	if opts.Album != "" {
		query.Set("album", opts.Album)
	}
	if opts.Artist != "" {
		query.Set("artist", opts.Artist)
	}
	if opts.Year != "" {
		query.Set("year", opts.Year)
	}
	return query.Encode()
}

// searchAppleMusic performs a music search query on the iTunes API and returns the search results.
func searchAppleMusic(query string) ([]Result, error) {
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
