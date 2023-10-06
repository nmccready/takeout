package mapper

import (
	"sort"

	"github.com/nmccready/takeout/src/internal/logger"
)

// nolint: deadcode,unused
var debug = logger.Spawn("mapper")

// Breaks up a map into chunks of the given size.
func ChunkBy[T any](items map[string]T, chunkSize int) (chunks []map[string]T) {
	var index int
	mapKeys := GetKeys(items)
	for index = 0; index < len(items); index += chunkSize {
		chunkMap := map[string]T{}
		for _, key := range mapKeys[index : index+chunkSize] {
			if key == "" {
				continue
			}
			chunkMap[key] = items[key]
		}
		chunks = append(chunks, chunkMap)
	}
	return chunks
}

func GetKeys[T any](items map[string]T) (keys []string) {
	for key := range items {
		keys = append(keys, key)
	}
	// sort keys to have reliable order, and easier to test
	sort.Strings(keys)
	return keys
}
