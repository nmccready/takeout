package model

import "strings"

type Tracker struct{}

type Track struct {
	Title          string
	Album          string
	Artist         string
	SupportArtists []string
	DurationSec    string
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

func toTrack(row []string) Track {
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
func (t Tracker) Parse(csv [][]string) (Tracks, TrackArtistAlbumMap) {
	tracks := Tracks{}
	trackMap := TrackArtistAlbumMap{}
	for ri, row := range csv {
		if ri == 0 {
			continue // skip header
		}
		track := toTrack(row)
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
