package model

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	gDebug "github.com/nmccready/go-debug"
	"github.com/nmccready/takeout/src/async"
	"github.com/nmccready/takeout/src/functional"
	"github.com/nmccready/takeout/src/json"
	"github.com/nmccready/takeout/src/slice"
)

type Job struct {
	Csv   string
	Paths []string
}

type JobResult struct {
	Err error
	Tracks
	ArtistAlbumMap
}

func (j JobResult) Error() error {
	return j.Err
}

func ParseMp3Glob(mp3Path string) (error, Tracks, ArtistAlbumMap) {

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

	var jobResults []JobResult

	err, jobResults = async.ProcessAsyncJobsByCpuNum[Job, JobResult](
		func(chunks int) []Job {
			return pathsToJobs(bytes, paths, chunks)
		},
		func(id int, job Job, jobResultChannel chan JobResult) {
			processPathChunks(id, job, jobResultChannel)
		},
	)

	if err != nil {
		return err, nil, nil
	}

	// merge together all Job Chunks / Maps etc.
	albumMap := ArtistAlbumMap{}
	tracks := Tracks{}
	for _, result := range jobResults {
		albumMap = albumMap.Merge(result.ArtistAlbumMap)
		tracks = append(result.Tracks, tracks...)
	}
	return nil, tracks, albumMap
}

func pathsToJobs(csv []byte, paths []string, chunks int) []Job {
	jobs := []Job{}
	pathChunks := slice.ChunkBy[string](paths, chunks)
	for _, pathsChunk := range pathChunks {
		jobs = append(jobs, Job{
			Paths: pathsChunk,
			Csv:   string(csv),
		})
	}
	return jobs
}

type TrackSearch struct {
	Basename     string
	Mp3FileName  string
	MetaReg      *TrackRegExpOpts
	OrigFilename string
}

func InitTrackSearch(origFilename, filePath string) *TrackSearch {
	basename := strings.ReplaceAll(origFilename, path.Ext(filePath), "")
	mp3FileName := cleanTrackMp3FileName(basename)
	metaReg := setupMetaPatterns(mp3FileName, titleMetaMatchStr)
	ts := TrackSearch{
		Basename:     basename,
		Mp3FileName:  mp3FileName,
		MetaReg:      metaReg,
		OrigFilename: origFilename,
	}
	return &ts
}

/*
Take a group of paths and out put them as a slice and map of tracks
map:

	Artists -> Album -> Songs
*/
func processPathChunks(id int, job Job, jobResultChannel chan JobResult) {
	tracks := Tracks{}
	trackMap := ArtistAlbumMap{}
	csv := job.Csv

	for _, filePath := range job.Paths {
		debug.Log("path: %s", filePath)
		err, track, origFilename := ParseId3ToTrack(filePath)

		ts := InitTrackSearch(origFilename, filePath)
		basename := ts.Basename
		mp3FileName := ts.Mp3FileName
		metaReg := ts.MetaReg

		debug.Log(gDebug.Fields{
			"regexStr":      metaReg.RegexStr,
			"titleRegexStr": metaReg.TitleRegexStr,
			"basename":      basename,
		})

		// partial id3 read / fix
		if err == nil && track.Title == "" &&
			track.Artist != "" &&
			track.Album != "" {
			track.Title = mp3FileName
			track.OrigFilename = origFilename
		}

		// NOTE: we can't use hashmap to title to csv rows due Google's Handling of titles
		// on file system not matching what they are in the csv / meta document
		// instead we must grep for matches to the csv rows (close enough to the title)

		// attempt to dequeue the csv to reduce matches
		if track.Title != "" { // id tag found remove it from csv
			track.removeFromCsv(csv)
		}

		// BEGIN FALLBACK TO METADATA FILE
		if err != nil || track.Title == "" {
			// utilize csv to derive the info via grep / regex
			var trackRef *Track
			trackRef, csv, err = getTrackFromCsvClean(origFilename, csv, metaReg)
			if err != nil {
				debug.Error("grepToTracks! %w", err)
				jobResultChannel <- JobResult{Err: err}
				return
			}
			track = *trackRef

			if track.Title == "" {
				err = fmt.Errorf("Missing Title from file %s", basename)
				debug.Error(err.Error())
				jobResultChannel <- JobResult{Err: err}
				return
			}
			if track.OrigFilename == "" {
				err = fmt.Errorf("Missing OriginalFile %s", basename)
				debug.Error(err.Error())
				jobResultChannel <- JobResult{Err: err}
				return
			}
		}
		debug.Log("track: %s", json.Stringify(track))
		tracks = append(tracks, track)
		trackMap.Add(&track)
	}
	jobResultChannel <- JobResult{Err: nil, Tracks: tracks, ArtistAlbumMap: trackMap}
}

func (track *Track) removeFromCsv(csv string) (string, error) {
	id3Reg := setupMetaPatterns(track.ToMetaRow(), titleId3MatchStr)
	id3Matches := id3Reg.TrackRegex.FindStringSubmatch(csv)
	debug.Log(gDebug.Fields{
		"id3Matches": id3Matches,
		"id3Reg":     *id3Reg,
	})
	if len(id3Matches) != 0 {
		csv = id3Reg.TrackRegex.ReplaceAllString(csv, "")
	}
	return csv, nil
}

func getTrackInfoFromCsv(origFilename, csv string, metaReg *TrackRegExpOpts) (*Track, error) {
	matches := metaReg.TrackRegex.FindAllStringSubmatch(csv, -1)
	debug.Log("metaReg: %+v", metaReg)
	debug.Log("matches: %s", json.Stringify(matches))
	if len(matches) == 0 {
		return nil, fmt.Errorf("No matches found for %s", metaReg.TrackRegex)
	}
	// grep tracks
	return grepToTracks(flattenMatches(matches), origFilename)
}

func getTrackFromCsvClean(origFilename, csv string, metaReg *TrackRegExpOpts) (track *Track, retCsv string, err error) {
	err = functional.LazyFn(
		func() error {
			track, err = getTrackInfoFromCsv(origFilename, csv, metaReg)
			return err
		},
		func() error {
			retCsv, err = track.removeFromCsv(csv)
			return err
		})
	if err != nil {
		return nil, "", err
	}
	return track, retCsv, nil
}

/*
google only allows the title whatever.mp3 to be so long and it gets cut off

So here we do a title header fudge factor to allow our meta to match.
*/
var titleMetaMatchStr = `[\w|)|(|\[|\]|\.|,|&|#|!]*"?,`
var titleId3MatchStr = `"?,` // more exact using to dequeue csv

type TrackRegExpOpts struct {
	RegexStr      string
	TitleRegexStr string
	TrackRegex    *regexp.Regexp
	TitleRegex    *regexp.Regexp
}

func setupMetaPatterns(mp3FileName, titleAppend string) *TrackRegExpOpts {
	opts := TrackRegExpOpts{}
	// grep basic raw file
	opts.RegexStr = `(?mi)^"?` + safeRegex(mp3FileName) + titleAppend + ".*"
	opts.TitleRegexStr = safeRegex(mp3FileName)

	opts.TrackRegex = regexp.MustCompile(opts.RegexStr)
	opts.TitleRegex = regexp.MustCompile(opts.TitleRegexStr)
	return &opts
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

func grepToTracks(matches []string, origFilename string) (*Track, error) {
	rows, err := grepToCsv(matches)
	if err != nil {
		return nil, err
	}
	// get file number if any to get match index
	index := 0
	numberMatches := fileNumberOnly.FindAllStringSubmatch(origFilename, -1)
	if len(numberMatches) != 0 {
		index, err = strconv.Atoi(numberMatches[0][1])
		if err != nil {
			return nil, err
		}
	}
	track := ToTrack(rows[index], origFilename)
	return &track, nil
}

var incrementedFileName = regexp.MustCompile(`(.*)\(\d+\)`)
var fileNumberOnly = regexp.MustCompile(`\((\d+)\)`)

/*
Title Names to Filenames considerations

It appears ' " * are substituted for _ in track file names

We need to make some potential matches to search the meta file
*/
func cleanTrackMp3FileName(filename string) string {
	return incrementedFileName.ReplaceAllString(filename, "$1")
}

var specialChars = []string{`\`, "(", ")", "[", "]", "#", "!", "$", "*", "+"}

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
