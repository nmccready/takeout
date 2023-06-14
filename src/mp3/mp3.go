package mp3

import (
	"github.com/bogem/id3v2"
)

// EncodeMetadata encodes the given metadata into the specified MP3 file.
func EncodeMetadata(filename string, metadata Metadata) error {
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
	tag.SetYear(metadata.Year)
	tag.SetGenre(metadata.Genre)

	// Save changes
	err = tag.Save()
	if err != nil {
		return err
	}

	return nil
}

func EncodeId(filename string, metadata Metadata) error {
	return EncodeMetadata(filename, metadata)
}

// Metadata represents the MP3 metadata.
type Metadata struct {
	Title  string
	Artist string
	Album  string
	Year   string
	Genre  string
}

// // Usage example:
// func main() {
// 	metadata := Metadata{
// 		Title:  "My Song",
// 		Artist: "John Doe",
// 		Album:  "My Album",
// 		Year:   "2023",
// 		Genre:  "Pop",
// 	}

// 	err := EncodeMetadata("path/to/myfile.mp3", metadata)
// 	if err != nil {
// 		panic(err)
// 	}
// }
