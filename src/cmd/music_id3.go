package cmd

import (
	"fmt"

	"github.com/nmccready/takeout/src/json"
	"github.com/nmccready/takeout/src/model"
	"github.com/spf13/cobra"
)

func init() {
	musicCmd.AddCommand(musicId3)
}

var musicId3 = &cobra.Command{
	Use:   "id3",
	Short: "read an mp3 file and output its id3 content",
	RunE: func(cmd *cobra.Command, args []string) error {
		mp3 = args[0]

		if mp3 == "" {
			panic("filepath and file name of music meta is required")
		}

		err, track, _ := model.ParseId3ToTrack(mp3)

		if err != nil {
			return err
		}

		fmt.Println(json.StringifyPretty(track))
		return nil
	},
}
