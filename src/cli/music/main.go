package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/nmccready/takeout/src/music"
)

func main() {
	// Check if a search query is provided as a command-line argument
	if len(os.Args) < 2 {
		fmt.Println("Please provide a search query.")
		return
	}

	// Perform the search
	query := strings.Join(os.Args[1:], " ")
	results, err := music.Search(music.SearchOpts{
		Title: query,
	})
	if err != nil {
		fmt.Printf("An error occurred during the search: %s\n", err)
		return
	}

	// Print the search results
	fmt.Printf("Search results for '%s':\n", query)
	for i, result := range results {
		fmt.Printf("%d. %s - %s (%s) [Source: %s]\n", i+1, result.Title, result.Artist, result.ReleaseYear, result.Source)
	}
}
