package gomicrobatcher

import (
	"sync"
	"time"
)

// JobProcessor defines the interface for processing jobs.
type JobProcessor interface {
	Process(jobs []interface{}) []interface{}
}

// Batcher represents the batching system.
type Batcher interface {
	AddJob(job interface{}) (JobResult, error)
	Stop()
}

// JobResult represents the result of a job.
type JobResult interface {
	GetValue() interface{}
}

// GenericBatch represents a batch of jobs.
type GenericBatch struct {
	values []jobValue
	timer  *timerWithStop
}

// genericBatcher is the implementation of the Batcher interface.
type genericBatcher struct {
	processor JobProcessor
	batchSize int
	linger    time.Duration
	pending   *GenericBatch
	mu        sync.Mutex
	shutdown  bool
}

// jobValue represents a job and its result channel.
type jobValue struct {
	value    interface{}
	resultCh chan interface{}
}

// timerWithStop is a timer with additional control.
type timerWithStop struct {
	timer  *time.Timer
	stopCh chan bool
}
