package workers

import (
	"sync"

	"github.com/go-openapi/errors"
	"github.com/nmccready/takeout/src/internal/logger"
)

var debug = logger.Spawn("workers")

const DEFAULT_CHUNK_SIZE = 5

// generic Response using generics
type Response[T any] struct {
	Err     error
	Payload T
}

type WorkerOpts[T any] struct {
	Item      []T
	RespChan  chan *Response[T]
	WaitGroup *sync.WaitGroup
}

type WorkerFunc[T any] func(*WorkerOpts[T])

type WorkerIdFunc[T any] func(*T) string

type RunWorkerOpts[T any] struct {
	Work []T
	WorkerIdFunc[T]
	WorkerFunc[T]
	DoBatchErrors bool
	ChunkSize     int
}

var dbgRw = debug.Spawn("RunWorkers")

func RunWorkers[T any](opts *RunWorkerOpts[T]) ([]T, error) {
	if opts.ChunkSize <= 0 {
		opts.ChunkSize = DEFAULT_CHUNK_SIZE
	}
	dbgRw.Log("begin")

	// based off https://golangcode.com/errors-in-waitgroups/
	responses := []T{}
	respChannel := make(chan *Response[T])
	waiter := sync.WaitGroup{}

	chunks := makeChunks(opts.Work, opts.ChunkSize)

	dbgRw.Log("opts.Work length: %d", len(opts.Work))
	dbgRw.Log("chunks length: %d", len(chunks))
	// run work chunks in parallel
	for _, chunk := range chunks {
		waiter.Add(1)
		go opts.WorkerFunc(&WorkerOpts[T]{Item: chunk, RespChan: respChannel, WaitGroup: &waiter})
	}

	batchErrors := []error{}

	// wait on all responses to be collected
	go func() {
		dbgRw.Log("Main: Waiting for workers to finish")
		waiter.Wait()
		close(respChannel)
		dbgRw.Log("Main: Wait Done,Closed respChannel")
	}()

	// collect responses
	for resp := range respChannel {
		dbgRw.Log("Main: Received response")
		if resp.Err != nil {
			if !opts.DoBatchErrors {
				dbgRw.Error("Main: Received response with error: %s", resp.Err)
				return responses, resp.Err
			}
			dbgRw.Warn("Main: Received worker batch error")
			batchErrors = append(batchErrors, resp.Err)
		}
		responses = append(responses, resp.Payload)
	}
	dbgRw.Log("responses length: %d", len(responses))
	if len(batchErrors) == 0 {
		return responses, nil
	}
	return responses, errors.CompositeValidationError(batchErrors...)

}

func makeChunks[T any](work []T, chunkSize int) [][]T {
	chunks := [][]T{}
	for i := 0; i < len(work); i += chunkSize {
		end := i + chunkSize
		if end > len(work) {
			end = len(work)
		}
		chunks = append(chunks, work[i:end])
	}
	return chunks
}
