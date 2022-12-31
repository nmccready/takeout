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
	musicCmd.Flags().StringVarP(&meta, "meta", "m", "", "filepath and file name of music meta to reorganize (required)")
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

		fmt.Printf("meta: %s\n", meta)

		file, err := os.Open(meta)
		if err != nil {
			return err
		}
		reader := csv.NewReader(file)
		records, _ := reader.ReadAll()

		tracks := model.Tracks{}.Parse(records)

		fmt.Println(model.ToJSONPretty(tracks))
		return nil
	},
}
