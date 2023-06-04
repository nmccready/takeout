package model

import (
	"fmt"
	"os"
	"strings"

	"github.com/nmccready/takeout/src/json"
	_strings "github.com/nmccready/takeout/src/strings"

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
type AlbumMap map[string]Tracks
type AlbumSongsMap map[string]Songs

/*
	"Tool": {
		"Fear Innoculumn": [
			"Tempest.mp3"
		]
	}
*/
type TrackArtistAlbumMap map[string]AlbumMap

/*
Appears to be broken, need to find a reliable merge library, generics would be amazing
for deepMerge
*/
func (m1 TrackArtistAlbumMap) Merge(m2 TrackArtistAlbumMap) TrackArtistAlbumMap {
	merged := make(TrackArtistAlbumMap)
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

/*
Appears to be broken, need to find a reliable merge library, generics would be amazing
for deepMerge
*/
func (m1 AlbumMap) Merge(m2 AlbumMap) AlbumMap {
	merged := make(AlbumMap)
	for k, v := range m1 {
		merged[k] = v
	}

	for key, value := range m2 {
		tracks, hasValue := merged[key]
		if hasValue {
			merged[key] = tracks.Merge(value)
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

func toTrack(row []string, origFilename string) Track {
	debug.Log("row: %s", json.Stringify(row))
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
			trackMap[track.Artist] = AlbumMap{track.Album: {track}}
			continue
		}
		if trackMap[track.Artist][track.Album] == nil {
			trackMap[track.Artist][track.Album] = Tracks{track}
			continue
		}
		trackMap[track.Artist][track.Album] = append(trackMap[track.Artist][track.Album], track)
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
		trackMap[artist] = AlbumMap{track.Album: {*track}}
		return
	}
	if trackMap[artist][track.Album] == nil {
		trackMap[artist][track.Album] = Tracks{*track}
		return
	}
	trackMap[artist][track.Album] = append(trackMap[artist][track.Album], *track)
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

func (tMap TrackArtistAlbumMap) Analysis() string {
	artists := len(tMap)
	albums := 0
	for _, aMap := range tMap {
		albums += len(aMap)
	}
	return fmt.Sprintf("%d artists, %d albums", artists, albums)
}

func (tMap TrackArtistAlbumMap) ToAlbumSongsMap() AlbumSongsMap {
	albumSongsMap := AlbumSongsMap{}
	for _, albumMap := range tMap {
		for album, tracks := range albumMap {
			albumSongsMap[album] = append(albumSongsMap[album], tracks.ToSongs()...)
		}
	}
	return albumSongsMap
}

func (track Track) ToMetaRow() string {
	return fmt.Sprintf("%s,%s,%s", track.Title, track.Album, track.Artist)
}
