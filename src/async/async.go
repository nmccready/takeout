package async

import (
	"runtime"
	"sync"

	"github.com/nmccready/takeout/src/internal/logger"
)

// nolint
var debug = logger.Spawn("async")

type IJobResult interface {
	Error() error
}

// Process Async Jobs via Wait group and Channel via Max CPU
func ProcessAsyncJobs[J interface{}, R IJobResult](
	numWorkers int,
	maxProcessors int,
	getJobs func(chunks int) []J,
	doWork func(id int, job J, jobResultChannel chan R),
) (err error, jobResults []R) {

	runtime.GOMAXPROCS(maxProcessors)

	var wg sync.WaitGroup
	wg.Add(numWorkers)
	jobChannel := make(chan J)

	jobs := getJobs(numWorkers)
	jobResultChannel := make(chan R, len(jobs))

	var worker = func(id int, wg *sync.WaitGroup, jobChannel chan J, jobResultChannel chan R) {
		defer wg.Done()
		for job := range jobChannel {
			doWork(id, job, jobResultChannel)
		}
	}

	for i := 0; i < numWorkers; i++ {
		go worker(i, &wg, jobChannel, jobResultChannel)
	}

	// Send jobs to worker
	for _, job := range jobs {
		jobChannel <- job
	}
	close(jobChannel)
	wg.Wait()
	close(jobResultChannel)

	// Receive job results from workers
	for result := range jobResultChannel {
		if result.Error() != nil {
			return result.Error(), nil
		}
		jobResults = append(jobResults, result)
	}

	return nil, jobResults
}

func ProcessAsyncJobsByCpuNum[J interface{}, R IJobResult](
	getJobs func(chunks int) []J,
	doWork func(id int, job J, jobResultChannel chan R),
) (err error, jobResults []R) {
	cpuNum := runtime.NumCPU()
	return ProcessAsyncJobs[J, R](cpuNum, cpuNum, getJobs, doWork)
}
