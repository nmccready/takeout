package cmd

import (
	"fmt"

	"github.com/nmccready/takeout/src/oauth2"
	"github.com/spf13/cobra"
)

func init() {
	musicCmd.AddCommand(musicDeezerToken)
}

// nolint
var musicDeezerToken = &cobra.Command{
	Use:   "deezer",
	Short: "search for music via deezer or ITunes",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := oauth2.GetDeezerToken()
		if err != nil {
			return err
		}
		fmt.Printf("token: %s\n", token.AccessToken)
		return nil
	},
}
