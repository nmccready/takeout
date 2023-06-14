package model

import (
	"os"
	"path"
	"runtime"
	"testing"

	_path "github.com/nmccready/takeout/src/path"
	"github.com/stretchr/testify/assert"
)

/*
Load a CSV fixture with some matching track details and
ensure that the tracks are parsed correctly from csv.

Duplicate filenames are index via 0 (Empty String) (1), (2), etc.

	Promises.mp3 - 0			csv first match
	Promises(1).mp3 - 1 	csv 2nd match
	Promises(2).mp3 - 2 	csv 3nd match
*/
func Test_getTrackInfoFromCsv(t *testing.T) {
	__dirname := _path.DirnameForce(_path.FilenameForce(runtime.Caller(0)))
	csvPath := path.Join(__dirname, "fixtures", "test.csv")
	//load csv file
	bytes, err := os.ReadFile(csvPath)

	assert.Nil(t, err)

	tests := []struct {
		OrigFilename string
		Track        *Track
	}{
		{
			OrigFilename: "Promises.mp3",
			Track: &Track{
				Title:        "Promises",
				Album:        "Blacc Hollywood (Deluxe Version)",
				Artist:       "Wiz Khalifa",
				DurationSec:  "210",
				OrigFilename: "Promises.mp3",
			},
		},
		{
			OrigFilename: "Promises(1).mp3",
			Track: &Track{
				Title:        "Promises",
				Album:        "Still Alive & Well?",
				Artist:       "Megadeth",
				DurationSec:  "269",
				OrigFilename: "Promises(1).mp3",
			},
		},
		{
			OrigFilename: "Promises(2).mp3",
			Track: &Track{
				Title:        "Promises",
				Album:        "The World Needs a Hero",
				Artist:       "Megadeth",
				DurationSec:  "269",
				OrigFilename: "Promises(2).mp3",
			},
		},
	}

	for _, test := range tests {
		ts := InitTrackSearch(test.OrigFilename, test.OrigFilename)
		track, err := getTrackInfoFromCsv(ts.OrigFilename, string(bytes), ts.MetaReg)
		assert.Nil(t, err, "found track")
		assert.Equal(t, test.Track, track, "track matches")
	}
}
