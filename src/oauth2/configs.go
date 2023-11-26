package oauth2

import (
	"fmt"

	"github.com/nmccready/takeout/src/os"
	"github.com/nmccready/oauth2"
)

type (
	RedirectOpts struct {
		Port string
		Base string
		Path string
	}
)

func ConfigDeezer(opts *RedirectOpts) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.GetRequiredEnv("DEEZER_APPLICATION_ID"),
		ClientSecret: os.GetRequiredEnv("DEEZER_SECRET_KEY"),
		RedirectURL: fmt.Sprintf(
			"http://%s:%s/%s", // https://redirectmeto.com/
			opts.Base,
			opts.Port,
			opts.Path,
		),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://connect.deezer.com/oauth/auth.php",
			TokenURL: "https://connect.deezer.com/oauth/access_token.php",
		},
		Scopes: []string{"basic_access"},
	}
}
