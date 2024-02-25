package gomicrobatcher

import (
	"errors"
	"log"
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
	frequency    time.Duration
	unresolved   *GenericBatch
	mu        sync.Mutex
	shutdown  bool
}

// jobValue represents a job and its result channel.
type jobValue struct {
	value    interface{}
	resultCh chan interface{}
}

// timerWithStop is a timer with and an additional channel to indicate that it needs to stop
type timerWithStop struct {
	timer  *time.Timer
	stopCh chan bool
}
// The job gets returned with an indicating channel 
type genericJobResult struct {
	ch    chan interface{}
	value interface{}
}

var ErrorMessage = errors.New("shutting down batcher")

func NewBatcher(customProcessor JobProcessor, size int, time time.Duration) Batcher {

	return &genericBatcher{
		processor: customProcessor, 
		batchSize: size, 
		frequency: time, 
		mu: sync.Mutex{},
		shutdown: false,
		unresolved: nil,	
	}
}

/*
In summary, this function adds a job to the batching system. 
If a batch is not already unresolved, it creates a new batch with the current job. 
If a batch is unresolved, it adds the job to the existing batch. 
If the batch size is reached, it triggers the execution of the batch. 
The function returns a JobResult instance, which can be used to retrieve the result of the job.
*/
func (b *genericBatcher) AddJob(job interface{}) (JobResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.shutdown {
		return nil, ErrorMessage
	}

	resultCh := make(chan interface{}, 1)
	jobVal := jobValue{value: job, resultCh: resultCh}

	if b.unresolved == nil {
		b.unresolved = b.newBatch(jobVal)
		return &genericJobResult{ch: resultCh}, nil
	}

	b.unresolved.values = append(b.unresolved.values, jobVal)

	if len(b.unresolved.values) >= b.batchSize {
		b.unresolved.stop()
		b.unresolved = nil
	}

	return &genericJobResult{ch: resultCh}, nil
}


// this time stops the batch of jobs
func (b *GenericBatch) stop() {
	b.timer.stop()
}

// stop the actual timer 
func (t *timerWithStop) stop() {
	t.stopCh <- true
	t.timer.Stop()
}

//helps to get the result of a single job when it is completed
func (j *genericJobResult) GetValue() interface{} {
	if j.value == nil {
		j.value = <-j.ch
		close(j.ch)
	}
	return j.value
}

func (b *genericBatcher) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.shutdown = true
	if b.unresolved != nil {
		b.unresolved.stop()
		b.unresolved = nil
	}
}

func newTimerWithStop(d time.Duration) timerWithStop {
	t := time.NewTimer(d)
	return timerWithStop{
		timer:t ,
		stopCh: make(chan bool, 1), // buffered channel
		
	}
}

func (b *genericBatcher) newBatch(firstJob jobValue) *GenericBatch {
	timer := newTimerWithStop(b.frequency)

	currentBatch := GenericBatch{
		values: []jobValue{firstJob},
		timer:  &timer,
	}

	go func() {
		select {
		case <-timer.stopCh:
		case <-timer.timer.C:
			log.Print("time expired")
			b.mu.Lock()
			b.unresolved = nil
			b.mu.Unlock()
		}

		b.executeBatch(currentBatch.values)
	}()

	return &currentBatch
}

func (b *genericBatcher) executeBatch(jobValues []jobValue) {
	values := make([]interface{}, len(jobValues))
	for i, jv := range jobValues {
		values[i] = jv.value
	}

	results := b.processor.Process(values)

	if len(results) != len(values) {
		panic("batch processor must produce a result for each value.")
	}

	for i, r := range results {
		jv := jobValues[i]
		select {
		case jv.resultCh <- r:
		default:
			log.Print("result channel closed or full")
		}
	}
}