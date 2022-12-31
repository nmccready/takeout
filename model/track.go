package model

import (
	"fmt"
	"os"
	osPath "path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bogem/id3v2/v2"
)

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

func (t Track) ParseId3(mp3FileName string) (error, Track) {
	tag, err := id3v2.Open(mp3FileName, id3v2.Options{Parse: true})
	if err != nil {
		return err, Track{}
	}
	defer tag.Close()
	t.Artist = tag.Artist()
	t.Album = tag.Album()
	t.Title = tag.Title()
	return nil, t
}

func (t Tracker) ParseMp3Glob(mp3Path string) (error, Tracks, TrackArtistAlbumMap) {
	tracks := Tracks{}
	trackMap := TrackArtistAlbumMap{}

	paths, err := filepath.Glob(mp3Path + "/*.mp3")
	if err != nil {
		return err, nil, nil
	}

	csvPaths, err := filepath.Glob(mp3Path + "/*.csv")
	if err != nil {
		return err, nil, nil
	}
	csvPath := csvPaths[0] // fallback if id3 fails

	bytes, err := os.ReadFile(csvPath)

	if err != nil {
		return err, nil, nil
	}

	csv := string(bytes)

	for _, path := range paths {
		debug.Log("path: %s", path)
		err, track := Track{}.ParseId3(path)
		if err != nil {
			// utilize csv to derive the info via grep / regex
			mp3FileNames := cleanTrackMp3FileName(strings.Split(osPath.Base(path), ".")[0])
			debug.Log("mp3FileNames: %s", ToJSON(mp3FileNames))
			var matches [][]string
			for _, mp3FileName := range mp3FileNames {
				trackRegex := regexp.MustCompile(safeRegex(".*" + mp3FileName + ".*"))
				matches = trackRegex.FindAllStringSubmatch(csv, -1)
				debug.Log("matches: %+v", matches)
				if len(matches) > 0 {
					break
				}
			}
			if len(matches) == 0 {
				debug.Error("Cannot Resolve Metadata!")
				return err, nil, nil
			}
			track = toTrack(strings.Split(matches[0][0], ","))
		}
		tracks = append(tracks, track)
		trackMap.Add(&track)
	}

	return nil, tracks, trackMap
}

func (trackMap TrackArtistAlbumMap) Add(track *Track) {
	if trackMap[track.Artist] == nil {
		trackMap[track.Artist] = AlbumMap{track.Album: {track.Title}}
		return
	}
	if trackMap[track.Artist][track.Album] == nil {
		trackMap[track.Artist][track.Album] = Songs{track.Title}
		return
	}
	trackMap[track.Artist][track.Album] = append(trackMap[track.Artist][track.Album], track.Title)
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

var incrementedFileName = regexp.MustCompile("(.*)\\(\\d\\)")

/*
	Title Names to Filenames considerations

	It appears ' " * are substituted for _ in track file names

	We need to make some potential matches to search the meta file
*/
func cleanTrackMp3FileName(filename string) []string {
	single := incrementedFileName.ReplaceAllString(strings.ReplaceAll(filename, "_", "'"), "$1")
	double := incrementedFileName.ReplaceAllString(strings.ReplaceAll(filename, "_", `"`), "$1")
	return []string{single, double}
}

// escape ( or )
func safeRegex(filename string) string {
	filename = strings.ReplaceAll(filename, "(", `\(`)
	filename = strings.ReplaceAll(filename, ")", `\)`)
	return filename
}
