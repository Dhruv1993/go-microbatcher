# go-microbatcher

Batching Library for Go

The Batching Library is a Go package designed to streamline the processing of jobs by efficiently batching them. It provides a simple interface for submitting jobs, allowing them to be processed in batches based on configurable parameters such as batch size, linger duration, and frequency.

Key Features:

Job Batching: Submit jobs to the library, and they will be processed in batches to improve efficiency.
Configurable Parameters: Adjust batch size, and processing frequency dynamically to suit your application's needs.
Graceful Shutdown: Safely shut down the batching system, ensuring all pending jobs are processed before exit.
Easy Integration: Integrate the library seamlessly into your Go applications to enhance job processing efficiency.

````markdown
```go
// Import the library
import "github.com/Dhruv1993/batchinglib"


// Create a new Batcher instance
batcher := batchinglib.NewBatcher(yourJobProcessor)

// Set custom parameters
batcher.SetBatchSize(10)
batcher.SetFrequency(20 * time.Millisecond)
```
````
