package cmd

import (
	"fmt"

	"github.com/nmccready/takeout/src/json"
	"github.com/nmccready/takeout/src/music"
	"github.com/spf13/cobra"

	"github.com/nmccready/takeout/src/internal/logger"
)

var debug = logger.Spawn("cmd:music:search")

type MusicSearchFlags struct {
	Title  string
	Album  string
	Artist string
	Year   string
}

var searchFlags MusicSearchFlags = MusicSearchFlags{}

func init() {
	musicSearch.Flags().StringVarP(&searchFlags.Title, "title", "t", "", "title of song")
	musicSearch.Flags().StringVarP(&searchFlags.Album, "album", "a", "", "album of song")
	musicSearch.Flags().StringVarP(&searchFlags.Artist, "artist", "r", "", "artist of song")
	musicSearch.Flags().StringVarP(&searchFlags.Year, "year", "y", "", "year of song")
	musicCmd.AddCommand(musicSearch)
}

// nolint
var musicSearch = &cobra.Command{
	Use:   "search",
	Short: "search for music via deezer or ITunes",
	RunE: func(cmd *cobra.Command, args []string) error {
		debug.Spawn("searchFlag").Log(json.StringifyPretty(searchFlags))
		results, err := music.Search(
			music.SearchOpts{
				Title:  searchFlags.Title,
				Album:  searchFlags.Album,
				Artist: searchFlags.Artist,
				Year:   searchFlags.Year,
			})

		if err != nil {
			return err
		}
		fmt.Println(json.StringifyPretty(results[0]))
		return nil
	},
}
