package cmd

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/nmccready/takeout/model"
	"github.com/spf13/cobra"
)

var meta string
var analyze bool
var doTrackMap bool

func init() {
	musicCmd.Flags().StringVarP(&meta, "meta", "m", "", "filepath and file name of music meta to reorganize (required)")
	musicCmd.Flags().BoolVarP(&analyze, "analyze", "a", false, "print tracks analysis")
	musicCmd.Flags().BoolVarP(&doTrackMap, "trackMap", "t", false, "print trackMap")
	err := musicCmd.MarkFlagRequired("meta")

	if err != nil {
		panic(err)
	}

	rootCmd.AddCommand(musicCmd)
}

var musicCmd = &cobra.Command{
	Use:   "music",
	Short: "Reorg the music files into their csc dir struct",
	RunE: func(cmd *cobra.Command, args []string) error {

		file, err := os.Open(meta)
		if err != nil {
			return err
		}
		reader := csv.NewReader(file)
		records, _ := reader.ReadAll()

		tracks, trackMap := model.Tracker{}.Parse(records)

		if doTrackMap {
			fmt.Println(model.ToJSONPretty(trackMap))
		}

		if analyze {
			fmt.Printf("Analysis: %s, %s\n", trackMap.Analysis(), tracks.Analysis())
		}
		return nil
	},
}
