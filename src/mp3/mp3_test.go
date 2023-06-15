package mp3

import (
	"os"
	"testing"

	"github.com/bogem/id3v2"
	"github.com/nmccready/takeout/src/json"
	"github.com/nmccready/takeout/src/model"
	"github.com/stretchr/testify/assert"
)

func createSampleMp3(t *testing.T, work func(*os.File, string)) error {
	filename := "sample.mp3"
	file, err := os.Create(filename)
	assert.NoError(t, err, "os.Create should not return an error")
	defer file.Close()
	work(file, filename)
	assert.NoError(t, os.Remove(filename), "os.Remove should not return an error")
	return nil
}

// Integration test
func TestEncodeMetadata(t *testing.T) {
	createSampleMp3(t, func(f *os.File, filename string) {
		// Encode metadata
		metadata := model.Track{
			Title:  "My Song",
			Artist: "John Doe",
			Album:  "My Album",
			Year:   2023,
			Genre:  "Pop",
		}
		err := EncodeMetadata(filename, &metadata)
		assert.NoError(t, err, "EncodeMetadata should not return an error")

		// Verify the encoded metadata
		tag, err := id3v2.Open(filename, id3v2.Options{Parse: true})
		assert.NoError(t, err, "id3v2.Open should not return an error")
		defer tag.Close()
		encodedMetadata := model.Id3V2TagToTrack(tag)

		assert.Equal(t, metadata, *encodedMetadata, "encoded metadata should match the expected values")
	})
}

func TestEncodeMetadataAgainstTrack(t *testing.T) {
	createSampleMp3(t, func(f *os.File, filename string) {
		// Encode metadata
		metadata := model.Track{
			Title:  "My Song",
			Artist: "John Doe",
			Album:  "My Album",
			Year:   2023,
			Genre:  "Pop",
		}
		err := EncodeMetadata(filename, &metadata)
		assert.NoError(t, err, "EncodeMetadata should not return an error")

		// Verify the encoded metadata
		err, track, origFilename := model.ParseId3ToTrack(filename)
		debug.Log("track: %s", json.StringifyPretty(track))
		debug.Log("metadata: %s", json.StringifyPretty(metadata))
		assert.NoError(t, err, "DecodeMetadata should not return an error")
		assert.Equal(t, filename, origFilename, "origFilename should match the expected value")

		assert.Equal(t, metadata, EncodedMetadataFromTrack(&track), "encoded metadata should match the expected values")
	})
}
