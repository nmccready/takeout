package cmd

import (
	"fmt"

	"github.com/nmccready/takeout/src/model"
	"github.com/spf13/cobra"
)

var analyze bool
var doTrackMap bool

func init() {
	musicCmd.Flags().BoolVarP(&analyze, "analyze", "a", false, "print tracks analysis")
	musicCmd.Flags().BoolVarP(&doTrackMap, "trackMap", "t", false, "print trackMap")

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
		err, _, trackMap := model.Tracker{}.ParseMp3Glob(mp3Path)

		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		if doTrackMap {
			fmt.Println(model.StringifyPretty(trackMap))
		}

		if analyze {
			// fmt.Printf("Analysis: %s, %s\n", trackMap.Analysis(), tracks.Analysis())
			fmt.Printf("Track Map Analysis: %s\n", trackMap.Analysis())
			// fmt.Printf("Track Analysis: %s\n", tracks.Analysis())
		}
		return nil
	},
}
