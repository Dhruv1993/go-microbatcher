package gomicrobatcher

import (
	"errors"
	"log"
	"sync"
	"time"
)

// JobProcessor defines the interface for processing jobs.
// We can do it via generics as well.
type JobProcessor interface {
	Process(jobs []interface{}) []interface{}
}

// Batcher represents the batching system. These methods will be exposed for the user
type Batcher interface {
	AddJob(job interface{}) (JobResult, error)
	Stop()
	SetBatchSize(size int)
	SetFrequency(time time.Duration)
}

// JobResult represents the result of a job.
type JobResult interface {
	GetValue() interface{}
}

// GenericBatch represents a batch of jobs.
type GenericBatch struct {
	values []JobValue
	timer  *TimerWithStop
}

// genericBatcher is the implementation of the Batcher interface.
type genericBatcher struct {
	processor    JobProcessor
	batchSize    int
	frequency    time.Duration
	unresolved   *GenericBatch
	mu           sync.Mutex  //since we want to update it concurrently from multiple goroutines, I have added a Mutex to synchronize access
	shutdown     bool
}

// JobValue represents a job and its result channel.
type JobValue struct {
	value    interface{}
	resultCh chan interface{}
}

// TimerWithStop is a timer with and an additional channel to indicate that it needs to stop
type TimerWithStop struct {
	timer  *time.Timer
	stopCh chan bool
}
// The job gets returned with an indicating channel 
type GenericJobResult struct {
	ch    chan interface{}
	value interface{}
}

var ErrorMessage = errors.New("shutting down batcher")


func NewBatcher(customProcessor JobProcessor) Batcher {
	return &genericBatcher{
		processor: customProcessor, 
		batchSize: 5, // default batch size
		frequency: 50*time.Millisecond, // default timeout
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
	jobVal := JobValue{value: job, resultCh: resultCh}

	if b.unresolved == nil {
		b.unresolved = b.newBatch(jobVal)
		return &GenericJobResult{ch: resultCh}, nil
	}

	b.unresolved.values = append(b.unresolved.values, jobVal)

	if len(b.unresolved.values) >= b.batchSize {
		b.unresolved.stop()
		b.unresolved = nil
	}

	return &GenericJobResult{ch: resultCh}, nil
}


// this time stops the batch of jobs
func (b *GenericBatch) stop() {
	b.timer.stop()
}

// stop the actual timer 
func (t *TimerWithStop) stop() {
	t.stopCh <- true
	t.timer.Stop()
}

//helps to get the result of a single job when it is completed
func (j *GenericJobResult) GetValue() interface{} {
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

func newTimerWithStop(d time.Duration) TimerWithStop {
	t := time.NewTimer(d)
	return TimerWithStop{
		timer:t ,
		stopCh: make(chan bool, 1), // buffered channel
		
	}
}

func (b *genericBatcher) newBatch(firstJob JobValue) *GenericBatch {
	timer := newTimerWithStop(b.frequency)

	currentBatch := GenericBatch{
		values: []JobValue{firstJob},
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

func (b *genericBatcher) executeBatch(JobValues []JobValue) {
	values := make([]interface{}, len(JobValues))
	for i, jv := range JobValues {
		values[i] = jv.value
	}

	results := b.processor.Process(values)

	for i, r := range results {
		jv := JobValues[i]
		select {
		case jv.resultCh <- r:
		default:
			log.Print("channel closed")
		}
	}
}

func (b *genericBatcher) SetFrequency(time time.Duration ) {
	b.frequency = time
}

func (b *genericBatcher) SetBatchSize(size int)  {
	b.batchSize = size
}