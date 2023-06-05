package model

import (
	"fmt"
	"os"

	_debug "github.com/nmccready/go-debug"
	"github.com/nmccready/takeout/src/async"
	"github.com/nmccready/takeout/src/json"
	"github.com/nmccready/takeout/src/mapper"
)

type AlbumMap map[string]Tracks
type AlbumSongsMap map[string]Songs

/*
	"Tool": {
		"Fear Innoculumn": [
			"Tempest.mp3"
		]
	}
*/
type ArtistAlbumMap map[string]AlbumMap

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

// warn about empty Artists??
func (trackMap ArtistAlbumMap) Add(track *Track) {
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

func (tMap ArtistAlbumMap) Analysis() string {
	artists := len(tMap)
	albums := 0
	for _, aMap := range tMap {
		albums += len(aMap)
	}
	return fmt.Sprintf("%d artists, %d albums", artists, albums)
}

func (tMap ArtistAlbumMap) ToAlbumSongsMap() AlbumSongsMap {
	albumSongsMap := AlbumSongsMap{}
	for _, albumMap := range tMap {
		for album, tracks := range albumMap {
			albumSongsMap[album] = append(albumSongsMap[album], tracks.ToSongs()...)
		}
	}
	return albumSongsMap
}

type SaveJob struct {
	MapChunk ArtistAlbumMap
}

type SaveJobResult struct {
	Err error
	ID  int
}

func (j SaveJobResult) Error() error {
	return j.Err
}

func (tMap ArtistAlbumMap) SaveChunk(src, dest string) error {
	// save each artist, album, and track
	for artist, albumMap := range tMap {
		for album, tracks := range albumMap {
			for _, track := range tracks {
				err := track.Save(SaveOpts{
					Src:    src,
					Dest:   dest,
					Artist: artist,
					Album:  album,
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Save All Tracks via ArtistAlbumMap structure using goroutines and works
// Async Save All
func (tMap ArtistAlbumMap) Save(src, dest string) error {
	var jobResults []SaveJobResult
	var err error
	dbg := debug.Spawn("Save")
	dbg.Log("in")

	// force make sure dest directory exists
	err = os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return err
	}

	err, jobResults = async.ProcessAsyncJobsByCpuNum[SaveJob, SaveJobResult](
		func(chunks int) []SaveJob {
			dbg.Log("chunks: %d", chunks)
			jobs := artistAlbumMapToSaveJobs(tMap, chunks)
			dbg.Spawn("jobs").Log(func() string { return json.StringifyPretty(jobs) })
			return jobs
		},
		func(id int, job SaveJob, jobResultChannel chan SaveJobResult) {
			// save all tracks in the chunk
			d := dbg.Spawn("job").Spawn("Save")
			d.Log(func() _debug.Fields {
				return _debug.Fields{
					"id":  id,
					"job": job,
				}
			})
			_err := job.MapChunk.SaveChunk(src, dest)
			if _err != nil {
				d.Error("%w", _err)
			}
			jobResultChannel <- SaveJobResult{Err: _err, ID: id}
		})

	if err != nil {
		return err
	}
	dbg.Log("SaveJobResults: %v", json.StringifyPretty(jobResults))
	return nil
}

func artistAlbumMapToSaveJobs(tMap ArtistAlbumMap, chunks int) []SaveJob {
	dbg := debug.Spawn("artistAlbumMapToSaveJobs")
	artistChunks := mapper.ChunkBy[AlbumMap](tMap, chunks)
	var jobs []SaveJob
	dbg.Log("artistChunks: %v", json.StringifyPretty(artistChunks))
	for _, artistChunk := range artistChunks {
		jobs = append(jobs, SaveJob{MapChunk: artistChunk})
	}
	dbg.Log("job len: %d", len(jobs))
	return jobs
}
