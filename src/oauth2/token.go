package oauth2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nmccready/takeout/src/internal/logger"
	"golang.org/x/oauth2"
)

var debug = logger.Spawn("oauth2")

type (
	HttpCallback func(w http.ResponseWriter, r *http.Request)

	OauthTokenResponse struct {
		*oauth2.Token
		Error error
	}

	ExchangeOpts struct {
		TokenResponseChan chan OauthTokenResponse
		AccessCodeChan    chan string
		RedirectOpts
	}
)

// getDeezerToken retrieves an OAuth2 token for the Deezer API.
func GetDeezerToken() (*oauth2.Token, error) {
	accessCodeChannel := make(chan string)
	tokenChannel := make(chan OauthTokenResponse)
	httpCb := genHandleRedirectCallback("code", accessCodeChannel)
	innerSO := ExchangeOpts{
		RedirectOpts: RedirectOpts{
			Port: "8080",
			Base: "", // hoping this is 0.0.0.0 ?
			Path: "deezer",
		},
		AccessCodeChan:    accessCodeChannel,
		TokenResponseChan: tokenChannel,
	}

	// Start HTTP server to handle Deezer API redirect callback
	http.HandleFunc("/"+innerSO.Path, httpCb)
	go http.ListenAndServe(fmt.Sprintf("%s:%s", innerSO.Base, innerSO.Port), nil)
	debug.Log("Listening on http://%s:%s/%s", innerSO.Base, innerSO.Port, innerSO.Path)
	go exchangeAccessCodeForToken(innerSO)
	debug.Log("Waiting for access code")

	payload := <-tokenChannel // hangs here
	debug.Log("Got access code")

	return payload.Token, payload.Error
}

func genHandleRedirectCallback(codeField string, accessCodeChannel chan string) HttpCallback {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract access code from query parameters
		// deezer returns the access code in the "code" query parameter
		accessCode := r.URL.Query().Get(codeField)
		// Send access code to channel
		accessCodeChannel <- accessCode
		// Respond to the user (e.g., redirect to a success page)
	}
}

func exchangeAccessCodeForToken(opts ExchangeOpts) {
	config := ConfigDeezer(&opts.RedirectOpts)
	debug.Log("config: %+v", config)

	// Wait for the access code
	config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	accessCode := <-opts.AccessCodeChan

	// Retrieve an access token
	token, err := config.Exchange(context.Background(), accessCode)
	opts.TokenResponseChan <- OauthTokenResponse{Token: token, Error: err}
}
