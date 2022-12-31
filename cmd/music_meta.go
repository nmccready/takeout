package cmd

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/nmccready/takeout/model"
	"github.com/spf13/cobra"
)

var meta string

func init() {
	musicMeta.Flags().BoolVarP(&analyze, "analyze", "a", false, "print tracks analysis")
	musicMeta.Flags().BoolVarP(&doTrackMap, "trackMap", "t", false, "print trackMap")
	musicCmd.AddCommand(musicMeta)
}

var musicMeta = &cobra.Command{
	Use:   "meta",
	Short: "read the meta file to compare to id3",
	RunE: func(cmd *cobra.Command, args []string) error {
		meta = args[0]

		if meta == "" {
			panic("filepath and file name of music meta is required")
		}

		file, err := os.Open(meta)
		if err != nil {
			return err
		}
		reader := csv.NewReader(file)
		records, _ := reader.ReadAll()

		tracks, trackMap := model.Tracker{}.ParseCsv(records)

		if doTrackMap {
			fmt.Println(model.ToJSONPretty(trackMap))
		}

		if analyze {
			fmt.Printf("Analysis: %s, %s\n", trackMap.Analysis(), tracks.Analysis())
		}
		return nil
	},
}
