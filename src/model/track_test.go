package model

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/nmccready/takeout/src/json"
	_path "github.com/nmccready/takeout/src/path"
	"github.com/stretchr/testify/assert"
)

func TestMergeAlbum(t *testing.T) {
	// assert for nil (good for errors)
	map1 := ArtistAlbumMap{
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
	map2 := ArtistAlbumMap{
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
	assert.Equal(t, ArtistAlbumMap{
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

	map3 := ArtistAlbumMap{
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
	expected := ArtistAlbumMap{
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

func TestTrackSave(t *testing.T) {
	// assert for nil (good for errors)
	track := Track{
		Title:        "Tempest",
		OrigFilename: "Tempest.mp3",
	}
	// "Tool": {
	// 	"Fear Innoculumn"
	// }
	// get fixture/takeout_dump directory path
	__dirname := _path.DirnameForce(_path.FilenameForce(runtime.Caller(0)))
	src := path.Join(__dirname, "fixtures", "takeout_dump")
	dest := path.Join(__dirname, "fixtures", "organized")
	track.Save(SaveOpts{
		Artist: "Tool",
		Album:  "Fear Innoculumn",
		Src:    src,
		Dest:   dest,
		DoCopy: true,
	})

	// assert song files in src, and dest are identical

	srcBytes, err := os.ReadFile(path.Join(src, "Tempest.mp3"))
	if err != nil {
		t.Errorf("Src Could not read file %v", err)
	}

	destBytes, err := os.ReadFile(path.Join(dest, "Tool", "Fear Innoculumn", "Tempest.mp3"))
	if err != nil {
		t.Errorf("Dest Could not read file %v", err)
	}

	assert.Equal(t, srcBytes, destBytes, "src and dest bytes are identical")
}
