package model

import (
	"testing"

	"github.com/nmccready/takeout/src/json"
	"github.com/stretchr/testify/assert"
)

func TestMergeAlbum(t *testing.T) {
	// assert for nil (good for errors)
	map1 := TrackArtistAlbumMap{
		"Tool": {
			"Fear Innoculumn": Tracks{
				Track{
					Title:        "Tempest",
					OrigFilename: "Tempest.mp3",
				},
				Track{
					Title:        "Invincible",
					OrigFilename: "Invincible.mp3",
				},
			},
		},
	}
	map2 := TrackArtistAlbumMap{
		"Tool": {
			"Fear Innoculumn": Tracks{
				Track{
					Title:        "Pneuma",
					OrigFilename: "Pneuma.mp3",
				},
			},
		},
		"Nine Inch Nails": {
			"Pretty Hate Machine": Tracks{
				Track{
					Title:        "Head Like A Hole",
					OrigFilename: "Head Like A Hole.mp3",
				},
			},
		},
	}

	merged := map1.Merge(map2)
	assert.Equal(t, TrackArtistAlbumMap{
		"Tool": {
			"Fear Innoculumn": Tracks{
				Track{
					Title:        "Tempest",
					OrigFilename: "Tempest.mp3",
				},
				Track{
					Title:        "Invincible",
					OrigFilename: "Invincible.mp3",
				},
				Track{
					Title:        "Pneuma",
					OrigFilename: "Pneuma.mp3",
				},
			},
		},
		"Nine Inch Nails": {
			"Pretty Hate Machine": Tracks{
				Track{
					Title:        "Head Like A Hole",
					OrigFilename: "Head Like A Hole.mp3",
				},
			},
		},
	}, merged, "basic merge one albumns per band")

	map3 := TrackArtistAlbumMap{
		"Tool": {
			"Laterlus": Tracks{
				Track{
					Title:        "Schism",
					OrigFilename: "Schism.mp3",
				},
			},
		},
		"Nine Inch Nails": {
			"The Downward Spiral": Tracks{
				Track{
					Title:        "March Of The Pigs",
					OrigFilename: "March Of The Pigs.mp3",
				},
				Track{
					Title:        "Closer",
					OrigFilename: "Closer.mp3",
				},
			},
		},
	}

	merged = merged.Merge(map3)
	debug.Spawn("test").Spawn("actual").Log(json.StringifyPretty(merged))
	expected := TrackArtistAlbumMap{
		"Tool": {
			"Fear Innoculumn": Tracks{
				Track{
					Title:        "Tempest",
					OrigFilename: "Tempest.mp3",
				},
				Track{
					Title:        "Invincible",
					OrigFilename: "Invincible.mp3",
				},
				Track{
					Title:        "Pneuma",
					OrigFilename: "Pneuma.mp3",
				},
			},
			"Laterlus": Tracks{
				Track{
					Title:        "Schism",
					OrigFilename: "Schism.mp3",
				},
			},
		},
		"Nine Inch Nails": {
			"Pretty Hate Machine": Tracks{
				Track{
					Title:        "Head Like A Hole",
					OrigFilename: "Head Like A Hole.mp3",
				},
			},
			"The Downward Spiral": Tracks{
				Track{
					Title:        "March Of The Pigs",
					OrigFilename: "March Of The Pigs.mp3",
				},
				Track{
					Title:        "Closer",
					OrigFilename: "Closer.mp3",
				},
			},
		},
	}
	debug.Spawn("test").Spawn("expected").Log(json.StringifyPretty(expected))
	assert.Equal(t, expected, merged, "basic merge one nested album's per band")
}
