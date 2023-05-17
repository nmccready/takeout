package model

import (
	"fmt"
	"os"
	"strings"

	_strings "github.com/nmccready/takeout/strings"

	id3 "github.com/dhowden/tag"
)

type Tracker struct{}

type Track struct {
	Title          string
	Album          string
	Artist         string
	SupportArtists []string
	DurationSec    string
	OrigFilename   string
}

type Tracks []Track
type Songs []string
type AlbumMap map[string]Songs

/*
	"Tool": {
		"Fear Innoculumn": [
			"Tempest.mp3"
		]
	}
*/
type TrackArtistAlbumMap map[string]AlbumMap

func (m1 TrackArtistAlbumMap) Merge(m2 TrackArtistAlbumMap) TrackArtistAlbumMap {
	merged := make(TrackArtistAlbumMap)
	for k, v := range m1 {
		merged[k] = v
	}

	for key, value := range m2 {
		album, hasValue := merged[key]
		if hasValue {
			album.Merge(value)
			continue
		}
		merged[key] = value
	}
	return merged
}

func (m1 AlbumMap) Merge(m2 AlbumMap) AlbumMap {
	merged := make(AlbumMap)
	for k, v := range m1 {
		merged[k] = v
	}

	for key, value := range m2 {
		songs, hasValue := merged[key]
		if hasValue {
			merged[key] = append(songs, value...)
			continue
		}
		merged[key] = value
	}
	return merged
}

func toTrack(row []string, origFilename string) Track {
	debug.Log("row: %s", ToJSON(row))
	track := Track{}
	track.Title = row[0]
	track.Album = row[1]
	artistsJoined := row[2]
	artists := strings.Split(artistsJoined, "/")
	track.Artist = artists[0]
	if len(artists) > 1 {
		track.SupportArtists = artists[1 : len(artists)-1]
	}
	track.DurationSec = row[3]
	return track
}

// note Empty String for Album or Artist is Unknown
func (t Tracker) ParseCsv(csv [][]string) (Tracks, TrackArtistAlbumMap) {
	tracks := Tracks{}
	trackMap := TrackArtistAlbumMap{}
	for ri, row := range csv {
		if ri == 0 {
			continue // skip header
		}
		track := toTrack(row, "")
		tracks = append(tracks, track)
		if trackMap[track.Artist] == nil {
			trackMap[track.Artist] = AlbumMap{track.Album: {track.Title}}
			continue
		}
		if trackMap[track.Artist][track.Album] == nil {
			trackMap[track.Artist][track.Album] = Songs{track.Title}
			continue
		}
		trackMap[track.Artist][track.Album] = append(trackMap[track.Artist][track.Album], track.Title)
	}
	return tracks, trackMap
}

/*
Main entry to try out different Id3 libs

Also tried:

"github.com/bogem/id3v2/v2" kinda works
"github.com/xonyagar/id3" v23 failures not much info but bad frames
*/
func (t Track) ParseId3(mp3FileName string) (error, Track) {
	// tag, err := id3v2.Open(mp3FileName, id3v2.Options{Parse: true})
	file, err := os.OpenFile(mp3FileName, os.O_RDONLY, 0666)
	if err != nil {
		debug.Error("failed to open file")
		return err, Track{}
	}
	// tag, err := id3.New(file)
	tag, err := id3.ReadFrom(file)
	if err != nil {
		return err, Track{}
	}
	defer file.Close()
	// defer tag.Close()
	// artists := tag.Artists()
	// t.Artist = artists[0]
	// t.Artist = tag.Artist()
	t.Artist = tag.AlbumArtist()
	t.Album = tag.Album()
	t.Title = tag.Title()
	// if len(artists) > 1 {
	// 	t.SupportArtists = artists[1 : len(artists)-1]
	// }
	return nil, t
}

func (t Track) GetArtistKey() string {
	artist := _strings.UnknownArtist
	if t.Artist != "" {
		artist = t.Artist
	}
	return artist
}

// warn about empty Artists??
func (trackMap TrackArtistAlbumMap) Add(track *Track) {
	artist := track.GetArtistKey()
	if trackMap[artist] == nil {
		trackMap[artist] = AlbumMap{track.Album: {track.Title}}
		return
	}
	if trackMap[artist][track.Album] == nil {
		trackMap[artist][track.Album] = Songs{track.Title}
		return
	}
	trackMap[artist][track.Album] = append(trackMap[artist][track.Album], track.Title)
}

func (tracks Tracks) Analysis() string {
	return fmt.Sprintf("%d songs", len(tracks))
}

func (tMap TrackArtistAlbumMap) Analysis() string {
	artists := len(tMap)
	albums := 0
	for _, aMap := range tMap {
		albums += len(aMap)
	}
	return fmt.Sprintf("%d artists, %d albums", artists, albums)
}

func (track Track) ToMetaRow() string {
	return fmt.Sprintf("%s,%s,%s", track.Title, track.Album, track.Artist)
}
