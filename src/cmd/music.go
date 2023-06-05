package cmd

import (
	"fmt"

	"github.com/nmccready/takeout/src/json"
	"github.com/nmccready/takeout/src/model"
	"github.com/nmccready/takeout/src/os"
	"github.com/spf13/cobra"
)

var analyze bool
var doTrackMap bool
var doTrackMapSimple bool
var dest string

func init() {
	musicCmd.Flags().BoolVarP(&analyze, "analyze", "a", false, "print tracks analysis")
	musicCmd.Flags().BoolVarP(&doTrackMap, "trackMap", "t", false, "print trackMap detailed")
	musicCmd.Flags().BoolVarP(&doTrackMapSimple, "trackMapSimple", "b", false, "print trackMap simple print")
	musicCmd.Flags().StringVarP(&dest, "save", "s", "", "absolute path of where to save tracks")

	rootCmd.AddCommand(musicCmd)
}

var musicCmd = &cobra.Command{
	Use:   "music",
	Short: "Reorg the music files into their csc dir struct",
	RunE: func(cmd *cobra.Command, args []string) error {
		mp3Path := args[0]

		if mp3Path == "" {
			panic("mp3Path required")
		}

		// err, tracks, trackMap := model.Tracker{}.ParseMp3Glob(mp3Path)
		err, _, trackMap := model.ParseMp3Glob(mp3Path)

		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		if doTrackMap {
			// simpler print for AlbumSongsMap
			fmt.Println(json.StringifyPretty(trackMap))
		}

		if doTrackMapSimple {
			// simpler print for AlbumSongsMap
			fmt.Println(json.StringifyPretty(trackMap.ToAlbumSongsMap()))
		}

		if dest != "" {
			os.ExitOnError(trackMap.Save(mp3Path, dest))
		}

		if analyze {
			// fmt.Printf("Analysis: %s, %s\n", trackMap.Analysis(), tracks.Analysis())
			fmt.Printf("Track Map Analysis: %s\n", trackMap.Analysis())
			// fmt.Printf("Track Analysis: %s\n", tracks.Analysis())
		}
		return nil
	},
}
