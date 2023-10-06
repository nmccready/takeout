package mp3

import (
	"strconv"

	"github.com/bogem/id3v2"
	"github.com/nmccready/takeout/src/internal/logger"
	"github.com/nmccready/takeout/src/model"
)

// nolint
var debug = logger.Spawn("mp3")

// EncodeMetadata encodes the given metadata into the specified MP3 file.
func EncodeMetadata(filename string, metadata *model.Track) error {
	tag, err := id3v2.Open(filename, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()

	tag.DeleteAllFrames()

	// Set new metadata
	tag.SetTitle(metadata.Title)
	tag.SetArtist(metadata.Artist)
	tag.SetAlbum(metadata.Album)
	tag.SetYear(strconv.Itoa(metadata.Year))
	tag.SetGenre(metadata.Genre)

	// Save changes
	err = tag.Save()
	if err != nil {
		return err
	}

	return nil
}

func EncodeId(filename string, metadata *model.Track) error {
	return EncodeMetadata(filename, metadata)
}

// Normalize what is defined from our Encoder to Match the more detailed Track
// parsed from id3
func EncodedMetadataFromTrack(track *model.Track) model.Track {
	return model.Track{
		Title:  track.Title,
		Artist: track.Artist,
		Album:  track.Album,
		Year:   track.Year,
		Genre:  track.Genre,
	}
}
