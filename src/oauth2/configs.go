package oauth2

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
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
		ClientID:     os.Getenv("DEEZER_APPLICATION_ID"),
		ClientSecret: os.Getenv("DEEZER_SECRET_KEY"),
		RedirectURL: fmt.Sprintf(
			"https://redirectmeto.com/http://%s:%s/%s",
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
