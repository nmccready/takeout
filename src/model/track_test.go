package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeAlbum(t *testing.T) {
	// assert for nil (good for errors)
	map1 := TrackArtistAlbumMap{
		"Tool": {
			"Fear Innoculumn": []string{
				"Tempest",
				"Invincible",
			},
		},
	}
	map2 := TrackArtistAlbumMap{
		"Tool": {
			"Fear Innoculumn": []string{
				"Pneuma",
			},
		},
		"Nine Inch Nails": {
			"Pretty Hate Machine": []string{
				"Head Like A Hole",
			},
		},
	}

	merged := map1.Merge(map2)
	assert.Equal(t, TrackArtistAlbumMap{
		"Tool": {
			"Fear Innoculumn": []string{
				"Tempest",
				"Invincible",
				"Pneuma",
			},
		},
		"Nine Inch Nails": {
			"Pretty Hate Machine": []string{
				"Head Like A Hole",
			},
		},
	}, merged, "basic merge one albumns per band")

	map3 := TrackArtistAlbumMap{
		"Tool": {
			"Laterlus": []string{
				"Schism",
			},
		},
		"Nine Inch Nails": {
			"The Downward Spiral": []string{
				"March Of The Pigs",
				"Closer",
			},
		},
	}

	merged = map1.Merge(map3)
	assert.Equal(t, TrackArtistAlbumMap{
		"Tool": {
			"Fear Innoculumn": []string{
				"Tempest",
				"Invincible",
				"Pneuma",
			},
			"Laterlus": []string{
				"Schism",
			},
		},
		"Nine Inch Nails": {
			"Pretty Hate Machine": []string{
				"Head Like A Hole",
			},
			"The Downward Spiral": []string{
				"March Of The Pigs",
				"Closer",
			},
		},
	}, merged, "basic merge one nested album's per band")
}
