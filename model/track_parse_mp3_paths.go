package model

import (
	"encoding/csv"
	"fmt"
	"os"
	osPath "path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	gDebug "github.com/nmccready/go-debug"
	"github.com/nmccready/takeout/slice"
)

func (t Tracker) ParseMp3Glob(mp3Path string) (error, Tracks, TrackArtistAlbumMap) {

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

	cpuNum := runtime.NumCPU()
	runtime.GOMAXPROCS(cpuNum)
	debug.Log("cpuNum: %d", cpuNum)
	var wg sync.WaitGroup
	wg.Add(cpuNum)
	jobChannel := make(chan Job)

	jobs := pathsToJobs(bytes, paths, cpuNum)
	jobResultChannel := make(chan JobResult, len(jobs))

	for i := 0; i < cpuNum; i++ {
		go worker(i, &wg, jobChannel, jobResultChannel)
	}

	// Send jobs to worker
	for _, job := range jobs {
		jobChannel <- job
	}
	close(jobChannel)
	wg.Wait()
	close(jobResultChannel)

	var jobResults []JobResult
	// Receive job results from workers
	for result := range jobResultChannel {
		jobResults = append(jobResults, result)
	}

	// merge together all Job Chunks / Maps etc.
	albumMap := TrackArtistAlbumMap{}
	tracks := Tracks{}
	for _, result := range jobResults {
		albumMap = albumMap.Merge(result.TrackArtistAlbumMap)
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

type Job struct {
	Csv   string
	Paths []string
}

type JobResult struct {
	Error error
	Tracks
	TrackArtistAlbumMap
}

func worker(id int, wg *sync.WaitGroup, jobChannel chan Job, jobResultChannel chan JobResult) {
	defer wg.Done()
	for job := range jobChannel {
		processPathChunks(id, job, jobResultChannel)
	}
}

/*
Take a group of paths and out put them as a slice and map of tracks
map:

	Artists -> Album -> Songs
*/
func processPathChunks(id int, job Job, jobResultChannel chan JobResult) {
	removedCsv := ""
	tracks := Tracks{}
	trackMap := TrackArtistAlbumMap{}
	csv := job.Csv

	for _, path := range job.Paths {
		debug.Log("path: %s", path)
		err, track := Track{}.ParseId3(path)

		basename := strings.ReplaceAll(osPath.Base(path), osPath.Ext(path), "")
		mp3FileName := cleanTrackMp3FileName(basename)

		var matches [][]string

		//nolint
		metaReg := setupMetaPatterns(mp3FileName, titleMetaMatchStr)

		debug.Log(gDebug.Fields{
			"regexStr":      metaReg.RegexStr,
			"titleRegexStr": metaReg.TitleRegexStr,
			"basename":      basename,
		})

		// partial id3 read / fix
		if err == nil && track.Title == "" &&
			track.Artist != "" &&
			track.Album != "" {
			track.Title = basename
		}

		// attempt to dequeue the csv to reduce matches
		if track.Title != "" {
			id3Reg := setupMetaPatterns(track.ToMetaRow(), titleId3MatchStr)
			id3Matches := id3Reg.TrackRegex.FindStringSubmatch(csv)
			debug.Log(gDebug.Fields{
				"id3Matches": id3Matches,
				"id3Reg":     *id3Reg,
			})
			if len(id3Matches) != 0 {
				lineToRemove := id3Matches[0]
				csv = id3Reg.TrackRegex.ReplaceAllString(csv, "")
				removedCsv += "\n" + lineToRemove
				debug.Log("lineToRemove: %s", lineToRemove)
				// debug.Log("removedCsv: %s", removedCsv)
				// debug.Error("sanity exit")
				// os.Exit(1)
			}
		}

		// BEGIN FALLBACK TO METADATA FILE
		if err != nil || track.Title == "" {
			// utilize csv to derive the info via grep / regex

			matches = metaReg.TrackRegex.FindAllStringSubmatch(csv, -1)
			debug.Log("matches: %s", ToJSON(matches))
			if len(matches) == 0 {
				debug.Error("Cannot Resolve Metadata!")
				jobResultChannel <- JobResult{Error: err}
				return
			}
			// grep tracks
			tracks, err := grepToTracks(flattenMatches(matches))
			if err != nil {
				debug.Error("grepToTracks! %w", err)
				jobResultChannel <- JobResult{Error: err}
				return
			}
			for _, _track := range tracks {
				debug.Log(gDebug.Fields{
					"_track":   _track,
					"regexStr": metaReg.RegexStr,
				})
				if metaReg.TitleRegex.MatchString(_track.Title) {
					track = _track
					break
				} else {
					debug.Log(gDebug.Fields{
						"titleRegexStr": metaReg.TitleRegexStr,
						"_track.Title":  _track.Title,
						"basename":      basename,
					})
				}
			}
			if track.Title == "" {
				err = fmt.Errorf("Missing Title from file %s", basename)
				debug.Error(err.Error())
				jobResultChannel <- JobResult{Error: err}
				return
			}
			matches = nil
		}
		debug.Log("track: %s", ToJSON(track))
		tracks = append(tracks, track)
		trackMap.Add(&track)
	}
	jobResultChannel <- JobResult{Error: nil, Tracks: tracks, TrackArtistAlbumMap: trackMap}
	return
}

/*
google only allows the title whatever.mp3 to be so long and it gets cut off

So here we do a title header fudge factor to allow our meta to match.
*/
var titleMetaMatchStr = `[\s|\w|)|(|\[|\]|\.|,|&|#|!]*"?,`
var titleId3MatchStr = `"?,` // more exact using to dequeue csv

type trackRegExpOpts struct {
	RegexStr      string
	TitleRegexStr string
	TrackRegex    *regexp.Regexp
	TitleRegex    *regexp.Regexp
}

func setupMetaPatterns(mp3FileName, titleAppend string) *trackRegExpOpts {
	opts := trackRegExpOpts{}
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

var incrementedFileName = regexp.MustCompile(`(.*)\(\d+\)`)

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