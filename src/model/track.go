package model

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/bogem/id3v2"
	id3 "github.com/dhowden/tag"
	"github.com/nmccready/takeout/src/json"
	_os "github.com/nmccready/takeout/src/os"
	_strings "github.com/nmccready/takeout/src/strings"
)

type Tracker struct{}

type Track struct {
	SupportArtists []string
	OrigFilename   string
	DurationSec    string

	// Begin id3 Metadata
	Format   id3.Format
	FileType id3.FileType
	Title    string

	Album       string
	Artist      string
	AlbumArtist string
	Composer    string
	Year        int
	Genre       string
	Picture     *id3.Picture
	Lyrics      string
	Comment     string

	Raw map[string]interface{}
}

func Id3MetadataToTrack(meta id3.Metadata) *Track {
	track := Track{}
	track.Format = meta.Format()
	track.FileType = meta.FileType()
	track.Title = meta.Title()
	track.Album = meta.Album()
	track.Artist = meta.Artist()
	track.AlbumArtist = meta.AlbumArtist()
	track.Composer = meta.Composer()
	track.Year = meta.Year()
	track.Genre = meta.Genre()
	track.Picture = meta.Picture()
	track.Lyrics = meta.Lyrics()
	track.Comment = meta.Comment()
	return &track
}

func Id3V2TagToTrack(tag *id3v2.Tag) *Track {
	year, _ := strconv.Atoi(tag.Year())
	return &Track{
		Title:  tag.Title(),
		Artist: tag.Artist(),
		Album:  tag.Album(),
		Year:   year,
		Genre:  tag.Genre(),
	}
}

type Tracks []Track
type Songs []string

var debugTrack = debug.Spawn("track")

/*
Appears to be broken, need to find a reliable merge library, generics would be amazing
for deepMerge
*/
func (m1 ArtistAlbumMap) Merge(m2 ArtistAlbumMap) ArtistAlbumMap {
	merged := make(ArtistAlbumMap)
	for k, v := range m1 {
		merged[k] = v
	}

	for key, value := range m2 {
		albums, hasValue := merged[key]
		if hasValue {
			merged[key] = albums.Merge(value)
			continue
		}
		merged[key] = value
	}
	return merged
}

// merge tracks array together
func (t1 Tracks) Merge(t2 Tracks) Tracks {
	tracks := Tracks{}
	tracks = append(tracks, t1...)
	tracks = append(tracks, t2...)
	return tracks
}

func ToTrack(csvRow []string, originalFilename string) Track {
	debugTrack.Log("row: %s", json.Stringify(csvRow))
	track := Track{}
	track.Title = csvRow[0]
	track.Album = csvRow[1]
	artistsJoined := csvRow[2]
	artists := strings.Split(artistsJoined, "/")
	track.Artist = artists[0]
	track.AlbumArtist = artists[0]
	if len(artists) > 1 {
		track.SupportArtists = artists[1 : len(artists)-1]
	}
	track.DurationSec = csvRow[3]
	track.OrigFilename = originalFilename
	return track
}

type OrigFilenameMap map[string]int // TrackName Map counter

// note Empty String for Album or Artist is Unknown
func ParseCsvToTracks(csv [][]string) (Tracks, ArtistAlbumMap) {
	trackNameCounter := OrigFilenameMap{}
	tracks := Tracks{}
	trackMap := ArtistAlbumMap{}
	for ri, row := range csv {
		if ri == 0 {
			continue // skip header
		}
		track := ToTrack(row, "") // need to figure out filename
		ctr, hasValue := trackNameCounter[track.Title]
		if !hasValue {
			track.OrigFilename = track.Title + ".mp3"
			trackNameCounter[track.Title] = 1
		} else {
			track.OrigFilename = fmt.Sprintf("%s(%d).mp3", track.Title, ctr)
			trackNameCounter[track.Title] = ctr + 1
		}
		tracks = append(tracks, track)
		if trackMap[track.AlbumArtist] == nil {
			trackMap[track.AlbumArtist] = AlbumMap{track.Album: {track}}
			continue
		}
		if trackMap[track.AlbumArtist][track.Album] == nil {
			trackMap[track.AlbumArtist][track.Album] = Tracks{track}
			continue
		}
		trackMap[track.AlbumArtist][track.Album] = append(trackMap[track.AlbumArtist][track.Album], track)
	}
	return tracks, trackMap
}

/*
Main entry to try out different Id3 libs

Also tried:

"github.com/bogem/id3v2/v2" kinda works
"github.com/xonyagar/id3" v23 failures not much info but bad frames
*/
func ParseId3ToTrack(mp3Path string) (error, Track, string) {
	// tag, err := id3v2.Open(mp3FileName, id3v2.Options{Parse: true})
	origFilename := path.Base(mp3Path)
	file, err := os.OpenFile(mp3Path, os.O_RDONLY, 0666)
	if err != nil {
		debugTrack.Error("failed to open file")
		return err, Track{}, origFilename
	}
	// tag, err := id3.New(file)
	tag, err := id3.ReadFrom(file)
	if err != nil {
		return err, Track{}, origFilename
	}
	defer file.Close()

	t := Id3MetadataToTrack(tag)
	t.OrigFilename = origFilename
	// if len(artists) > 1 {
	// 	t.SupportArtists = artists[1 : len(artists)-1]
	// }
	return nil, *t, origFilename
}

func (t Track) GetArtistKey() string {
	artist := _strings.UnknownArtist
	if t.Artist != "" {
		artist = t.Artist
	}
	return artist
}

func (tracks Tracks) Analysis() string {
	return fmt.Sprintf("%d songs", len(tracks))
}

func (tracks Tracks) ToSongs() Songs {
	songs := Songs{}
	for _, track := range tracks {
		songs = append(songs, track.Title)
	}
	return songs
}

func (track Track) ToMetaRow() string {
	return fmt.Sprintf("%s,%s,%s", track.Title, track.Album, track.Artist)
}

type SaveOpts struct {
	Src    string
	Dest   string
	Artist string
	Album  string
	DoCopy bool
}

var dbgTrackSave = debugTrack.Spawn("Save")

func (track Track) Save(opts SaveOpts) error {
	// save to file system
	destDir := fmt.Sprintf("%s/%s/%s", opts.Dest, opts.Artist, opts.Album)
	err := os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		return err
	}
	src := fmt.Sprintf("%s/%s", opts.Src, track.OrigFilename)
	dest := fmt.Sprintf("%s/%s", destDir, track.OrigFilename)
	dbgTrackSave.Log("src: %s, dest: %s", src, dest)

	if opts.DoCopy {
		return _os.Copy(src, dest)
	}
	return os.Rename(src, dest)
}
