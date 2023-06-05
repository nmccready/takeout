package mapper

import (
	"testing"

	"github.com/nmccready/takeout/src/json"
	"github.com/stretchr/testify/assert"
)

func TestChunkBy(t *testing.T) {
	mapToBreakUp := map[string]string{
		"a": "a", "b": "b", "c": "c",
		"d": "d", "e": "e", "f": "f", "g": "g"}

	chunks := ChunkBy(mapToBreakUp, 2)
	assert.Equal(t, 4, len(chunks))
	debug.Log(json.StringifyPretty(chunks))
	assert.Equal(t, []map[string]string{
		{
			"a": "a",
			"b": "b",
		},
		{
			"c": "c",
			"d": "d",
		},
		{
			"e": "e",
			"f": "f",
		},
		{
			"g": "g",
		},
	}, chunks)

}

func TestGetKeys(t *testing.T) {
	assert.Equal(t, []string{"a", "b"},
		GetKeys(map[string]string{"a": "a", "b": "b"}))
}
