package cmd

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/nmccready/takeout/src/json"
	"github.com/nmccready/takeout/src/model"
	"github.com/spf13/cobra"
)

var mp3 string

func init() {
	musicId3.Flags().BoolVarP(&analyze, "analyze", "a", false, "print tracks analysis")
	musicId3.Flags().BoolVarP(&doTrackMap, "trackMap", "t", false, "print trackMap")
	musicCmd.AddCommand(musicId3)
}

var musicMeta = &cobra.Command{
	Use:   "meta",
	Short: "read the meta file to compare to id3",
	RunE: func(cmd *cobra.Command, args []string) error {
		mp3 = args[0]

		if mp3 == "" {
			panic("filepath and file name of music meta is required")
		}

		file, err := os.Open(mp3)
		if err != nil {
			return err
		}
		reader := csv.NewReader(file)
		records, _ := reader.ReadAll()

		tracks, trackMap := model.Tracker{}.ParseCsv(records)

		if doTrackMap {
			fmt.Println(json.StringifyPretty(trackMap))
		}

		if analyze {
			fmt.Printf("Analysis: %s, %s\n", trackMap.Analysis(), tracks.Analysis())
		}
		return nil
	},
}
