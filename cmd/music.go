package cmd

import (
	"fmt"

	"github.com/nmccready/takeout/model"
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

		err, tracks, trackMap := model.Tracker{}.ParseMp3Glob(mp3Path)

		if err != nil {
			return err
		}

		if doTrackMap {
			fmt.Println(model.ToJSONPretty(trackMap))
		}

		if analyze {
			fmt.Printf("Analysis: %s, %s\n", trackMap.Analysis(), tracks.Analysis())
		}
		return nil
	},
}
