package model

import "strings"

type Track struct {
	Title          string
	Album          string
	Artist         string
	SupportArtists []string
	DurationSec    string
}

type Tracks []Track

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

func (tracks Tracks) Parse(csv [][]string) Tracks {
	for ri, row := range csv {
		if ri == 0 {
			continue // skip header
		}
		tracks = append(tracks, toTrack(row))
	}
	return tracks
}
