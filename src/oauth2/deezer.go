package oauth2

import (
	"fmt"
	"net/http"

	"github.com/nmccready/oauth2"
	"github.com/nmccready/takeout/src/os"
)

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

	maybeToken, _ := LoadToken(innerSO.Config.ClientID)
	if maybeToken != nil {
		return maybeToken, nil
	}

	// Start HTTP server to handle Deezer API redirect callback
	http.HandleFunc("/"+innerSO.Path, httpCb)
	// fix the below function so that it actually serves with a handler and not nil
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%s", innerSO.Port), nil)
		if err != nil {
			panic(err)
		}
	}()
	debug.Log("Listening on http://:%s/%s", innerSO.Port, innerSO.Path)
	go exchangeAccessCodeForToken(innerSO)
	debug.Log("Waiting for access code")

	payload := <-tokenChannel // hangs here
	debug.Log("Got access code")
	debug.Log("payload: %+v", payload)

	err := SaveToken(innerSO.Config.ClientID, payload.Token)
	if err != nil {
		return nil, err
	}

	return payload.Token, payload.Error
}
