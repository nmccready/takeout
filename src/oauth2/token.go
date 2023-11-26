package oauth2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nmccready/oauth2"
	"github.com/nmccready/takeout/src/internal/logger"
	"github.com/nmccready/takeout/src/os"
)

var debug = logger.Spawn("oauth2")

var STATIC_IP = os.GetRequiredEnv("STATIC_IP")
var DEEZER_PORT = os.GetRequiredEnv("DEEZER_PORT")

var CACHE_PATH = []string{"tmp", ".takeout", "ouath2_tokens.json"}

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
		AuthStyle      oauth2.AuthStyle
		AuthCodeOption oauth2.AuthCodeOption
	}
)

// Attempt to load the token from the cache file.
func LoadTokenFromCache() (*oauth2.Token, error) {
	token := &oauth2.Token{}
	// Load the token from the cache file
	err := os.LoadJSON(CACHE_PATH, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// getDeezerToken retrieves an OAuth2 token for the Deezer API.
func GetDeezerToken() (*oauth2.Token, error) {
	accessCodeChannel := make(chan string)
	tokenChannel := make(chan OauthTokenResponse)
	httpCb := genHandleRedirectCallback("code", accessCodeChannel)
	innerSO := ExchangeOpts{
		RedirectOpts: RedirectOpts{
			Port: DEEZER_PORT,
			Base: STATIC_IP,
			Path: "deezer",
		},
		AccessCodeChan:    accessCodeChannel,
		TokenResponseChan: tokenChannel,
		AuthStyle:         oauth2.AuthStyleAutoDetect,
		AuthCodeOption:    oauth2.AccessTypeOffline,
	}

	// Start HTTP server to handle Deezer API redirect callback
	http.HandleFunc("/"+innerSO.Path, httpCb)
	// fix the below function so that it actually serves with a handler and not nil
	go http.ListenAndServe(fmt.Sprintf(":%s", innerSO.Port), nil)
	debug.Log("Listening on http://:%s/%s", innerSO.Port, innerSO.Path)
	go exchangeAccessCodeForToken(innerSO)
	debug.Log("Waiting for access code")

	payload := <-tokenChannel // hangs here
	debug.Log("Got access code")
	debug.Log("payload: %+v", payload)

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
	safeConfig := *config
	safeConfig.ClientID = "********"
	safeConfig.ClientSecret = "********"
	debug.Log("config: %+v", safeConfig)

	// Wait for the access code
	url := config.AuthCodeURL("state", opts.AuthCodeOption)
	fmt.Printf("Visit the URL for the auth dialog: %v", url)
	accessCode := <-opts.AccessCodeChan

	debug.Log("accessCode: %s", accessCode)

	// Retrieve an access token
	config.Endpoint.AuthStyle = opts.AuthStyle
	token, err := config.Exchange(context.Background(), accessCode) // FAILING HERE
	// payload: {Token:<nil> Error:oauth2: cannot parse json: invalid character 'a' looking for beginning of value}
	opts.TokenResponseChan <- OauthTokenResponse{Token: token, Error: err}
}
