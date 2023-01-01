package model

import (
	"encoding/csv"
	"fmt"
	"os"
	osPath "path"
	"path/filepath"
	"regexp"
	"strings"

	id3 "github.com/dhowden/tag"
	gDebug "github.com/nmccready/go-debug"
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

func (t Tracker) ParseMp3Glob(mp3Path string) (error, Tracks, TrackArtistAlbumMap) {
	tracks := Tracks{}
	trackMap := TrackArtistAlbumMap{}

	paths, err := filepath.Glob(mp3Path + "/" + "*.mp3")
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
		// var err error
		// track := Track{}

		// if err != nil {
		// 	debug.Error(gDebug.Fields{"path": path})
		// 	return err, nil, nil
		// }

		// BEGIN FALLBACK TO METADATA FILE
		if err != nil || track.Title == "" {
			// utilize csv to derive the info via grep / regex
			basename := strings.ReplaceAll(osPath.Base(path), osPath.Ext(path), "")
			mp3FileNames := cleanTrackMp3FileName(basename)
			debug.Log("mp3FileNames: %s", ToJSON(mp3FileNames))
			var matches [][]string
			var regexStr string
			var trackRegex *regexp.Regexp
			// grep basic raw file
			for _, mp3FileName := range mp3FileNames {
				regexStr = ".*" + safeRegex(mp3FileName) + ".*"
				debug.Log(gDebug.Fields{
					"regexStr": regexStr,
					"basename": basename,
				})
				trackRegex = regexp.MustCompile(regexStr)
				matches = trackRegex.FindAllStringSubmatch(csv, -1)
				debug.Log("matches: %s", ToJSON(matches))
				if len(matches) > 0 {
					break
				}
			}
			if len(matches) == 0 {
				debug.Error("Cannot Resolve Metadata!")
				return err, nil, nil
			}
			// grep tracks
			tracks, err := grepToTracks(flattenMatches(matches))
			if err != nil {
				debug.Error("grepToTracks! %w", err)
				return err, nil, nil
			}
			for _, _track := range tracks {
				debug.Log(gDebug.Fields{
					"_track":   _track,
					"regexStr": regexStr,
				})
				if trackRegex.MatchString(_track.Title) {
					track = _track
					break
				}
			}
			if track.Title == "" {
				err = fmt.Errorf("Missing Title from file %s", basename)
				debug.Error(err.Error())
				return err, nil, nil
			}
			matches = nil
		}
		debug.Log("track: %s", ToJSON(track))
		tracks = append(tracks, track)
		trackMap.Add(&track)
	}

	return nil, tracks, trackMap
}

func flattenMatches(rootMatches [][]string) []string {
	flat := []string{}
	for _, matches := range rootMatches {
		flat = append(flat, matches...)
	}
	return flat
}

func grepToCsv(matches []string) ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(strings.Join(matches, "\n")))
	return reader.ReadAll()
}

func grepToTracks(matches []string) (Tracks, error) {
	tracks := Tracks{}
	rows, err := grepToCsv(matches)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		tracks = append(tracks, toTrack(r, ""))
	}
	return tracks, nil
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

var incrementedFileName = regexp.MustCompile(`(.*)\(\d+\)`)

/*
	Title Names to Filenames considerations

	It appears ' " * are substituted for _ in track file names

	We need to make some potential matches to search the meta file
*/
func cleanTrackMp3FileName(filename string) []string {
	underscore := incrementedFileName.ReplaceAllString(filename, "$1")
	// star := incrementedFileName.ReplaceAllString(strings.ReplaceAll(filename, "_", `\*`), "$1")
	// ampersand := incrementedFileName.ReplaceAllString(strings.ReplaceAll(filename, "_", "&"), "$1")
	// single := incrementedFileName.ReplaceAllString(strings.ReplaceAll(filename, "_", "'"), "$1")
	// double := incrementedFileName.ReplaceAllString(strings.ReplaceAll(filename, "_", `"`), "$1")
	// return []string{underscore, star, ampersand, single, double}
	return []string{underscore}
}

var specialChars = []string{"(", ")", "[", "]", "#", "!", "$", "*", "+"}

// escape special chars for regex
// handle / relax _ meaning as Google uses it for a lot
func safeRegex(filename string) string {
	for _, char := range specialChars {
		filename = strings.ReplaceAll(filename, char, `\`+char)
	}
	// google fudge factor (_ utilized for swearing, and all the following chars)
	// %% to escape fmt.Sprintf to single %
	filename = strings.ReplaceAll(filename, "_", `[%%|\*|&|'|"|/|:|?\|_]+`)
	return filename
}
