package oauth2

import (
	"context"
	"fmt"
	"net/http"

	_os "os"

	"github.com/nmccready/oauth2"
	"github.com/nmccready/takeout/src/internal/logger"
	"github.com/nmccready/takeout/src/os"
)

var debug = logger.Spawn("oauth2")

var _cachePath = []string{".takeout", "ouath2_tokens.json"}
var cacheFilename = _cachePath[len(_cachePath)-1]

type (
	HttpCallback func(w http.ResponseWriter, r *http.Request)

	OauthTokenResponse struct {
		*oauth2.Token
		Error error
	}

	ExchangeOpts struct {
		TokenResponseChan chan OauthTokenResponse
		AccessCodeChan    chan string
		Config            *oauth2.Config
		RedirectOpts
		AuthStyle      oauth2.AuthStyle
		AuthCodeOption oauth2.AuthCodeOption
	}
)

type TokenCache = map[string]*oauth2.Token

// Attempt to load the token from the cache file.
func BaseLoadTokenCache(clientId string, cachePath []string) (*TokenCache, error) {
	cache := TokenCache{}
	// Load the token from the cache file
	err := os.LoadJSON(cachePath, &cache)
	if err != nil {
		return nil, err
	}
	return &cache, nil
}

func LoadTokenCache(clientId string) (*TokenCache, error) {
	homeDir, err := _os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	paths := append(_cachePath, homeDir)
	return BaseLoadTokenCache(clientId, paths)
}

func BaseLoadToken(clientId string, cachePath []string) (*oauth2.Token, error) {
	cache, err := BaseLoadTokenCache(clientId, cachePath)
	if err != nil {
		return nil, err
	}
	if token, ok := (*cache)[clientId]; ok {
		return token, nil
	}
	return nil, fmt.Errorf("token not found in cache")
}

func LoadToken(clientId string) (*oauth2.Token, error) {
	return BaseLoadToken(clientId, _cachePath)
}

// getDeezerToken retrieves an OAuth2 token for the Deezer API.
func GetDeezerToken() (*oauth2.Token, error) {
	deezerPort := os.GetRequiredEnv("DEEZER_PORT")
	staticIp := os.GetRequiredEnv("STATIC_IP")
	accessCodeChannel := make(chan string)
	tokenChannel := make(chan OauthTokenResponse)
	httpCb := genHandleRedirectCallback("code", accessCodeChannel)
	redirectOpts := RedirectOpts{
		Port: deezerPort,
		Base: staticIp,
		Path: "deezer",
	}
	innerSO := ExchangeOpts{
		RedirectOpts:      redirectOpts,
		AccessCodeChan:    accessCodeChannel,
		TokenResponseChan: tokenChannel,
		AuthStyle:         oauth2.AuthStyleAutoDetect,
		AuthCodeOption:    oauth2.AccessTypeOffline,
		Config:            ConfigDeezer(&redirectOpts),
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
	config := opts.Config
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
