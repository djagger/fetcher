package pool

import "fetcher/pkg/services"

const (
	maxFetchPoolWorkers = 2
	poolWorkersBuffer   = 5
)

type FetchPool interface {
	Start()
	AddWorker(params *services.RequesterParams)
}

type fetchPool struct {
	workerInput chan services.RequesterParams
	maxWorkers  int
	requester   services.Requester
}

func (f *fetchPool) Start() {
	for i := 0; i < f.maxWorkers; i++ {
		go f.startWorker(f.workerInput)
	}
}

func (f *fetchPool) startWorker(in <-chan services.RequesterParams) {
	for inputParams := range in {
		// Get new job for worker and do request.
		f.requester.Do(inputParams)
	}
}

func (f *fetchPool) AddWorker(params *services.RequesterParams) {
	f.workerInput <- *params
}

func NewFetchPool(requester services.Requester) FetchPool {
	workerInput := make(chan services.RequesterParams, poolWorkersBuffer)

	return &fetchPool{
		workerInput: workerInput,
		maxWorkers:  maxFetchPoolWorkers,
		requester:   requester,
	}
}
