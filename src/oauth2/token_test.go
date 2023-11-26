package oauth2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test to load the cache file from the fixture cache.jsonc
func TestLoadTokenCache(t *testing.T) {
	clientId := "1234567890"
	token, err := BaseLoadToken(clientId, []string{"./", "fixture", "cache.json"})
	assert.Nil(t, err, "error is nil")
	assert.NotNil(t, token, "token is defined")

	assert.Equal(t, "abcdefg", token.AccessToken, "access token matches")
	assert.Equal(t, "Bearer", token.TokenType, "token type matches")
	assert.Equal(t, "hijklmn", token.RefreshToken, "refresh token matches")
}
