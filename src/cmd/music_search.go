package cmd

import (
	"github.com/nmccready/takeout/src/music"
	"github.com/spf13/cobra"
)

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
	Short: "search for music via deezer or Itunes",
	RunE: func(cmd *cobra.Command, args []string) error {
		music.Search(searchFlags.Title, searchFlags.Album, searchFlags.Artist, searchFlags.Year)
		return nil
	},
}
