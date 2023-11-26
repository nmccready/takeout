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

type TokenCache map[string]*oauth2.Token

func (tc TokenCache) IsEmpty() bool {
	return len(tc) == 0
}

// Attempt to load the token from the cache file.
func BaseLoadTokenCache(clientId string, cachePath []string) (TokenCache, error) {
	cache := TokenCache{}
	// Load the token from the cache file
	// nolint
	_ = os.LoadJSON(cachePath, &cache)
	return cache, nil
}

func LoadTokenCache(clientId string) (TokenCache, error) {
	paths, err := os.GetHomeDirWithPaths(_cachePath)
	if err != nil {
		return nil, err
	}
	return BaseLoadTokenCache(clientId, paths)
}

var loadTokenDbg = debug.Spawn("BaseLoadToken")

func BaseLoadToken(clientId string, cachePath []string) (*oauth2.Token, error) {
	loadTokenDbg.Log("cachePath: %s", cachePath)
	cache, err := BaseLoadTokenCache(clientId, cachePath)

	if err != nil {
		loadTokenDbg.Error("token found not for clientId: %s, err: %s", clientId, err.Error())
		return nil, err
	}
	if token, ok := cache[clientId]; ok {
		loadTokenDbg.Log("token found for clientId: %s", clientId)
		return token, nil
	}
	return nil, fmt.Errorf("token not found in cache")
}

func LoadToken(clientId string) (*oauth2.Token, error) {
	paths, err := os.GetHomeDirWithPaths(_cachePath)
	if err != nil {
		return nil, err
	}
	return BaseLoadToken(clientId, paths)
}

func BaseSaveToken(clientId string, token *oauth2.Token, cachePath []string) error {
	cache, err := BaseLoadTokenCache(clientId, cachePath)
	if err != nil {
		return err
	}

	cache[clientId] = token
	return os.SaveJSON(cachePath, cache)
}

var saveTokenDbg = debug.Spawn("SaveToken")

func SaveToken(clientId string, token *oauth2.Token) error {
	paths, err := os.GetHomeDirWithPaths(_cachePath)
	if err != nil {
		return err
	}
	saveTokenDbg.Log("saved clientId: %s", clientId)
	return BaseSaveToken(clientId, token, paths)
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
