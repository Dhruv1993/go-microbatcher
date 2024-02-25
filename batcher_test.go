package gomicrobatcher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// for test results only
type prefixProcessor struct {
	Prefix string
}

func (p *prefixProcessor) Process(jobs []interface{}) []interface{} {
	result := make([]interface{}, len(jobs))
	for i, v := range jobs {
		strValue, ok := v.(string)
		if !ok { // if not string then panic
			panic("prefixProcessor: input is not a string")
		}
		result[i] = p.Prefix + strValue
	}
	return result
}

func Test_WhenBatchSizeIsSame(t *testing.T) {
	assert := assert.New(t)
	jobBatcher := NewBatcher(&prefixProcessor{Prefix: "Processed - "})
	jobBatcher.SetBatchSize(2)
	jobBatcher.SetFrequency(20*time.Millisecond)

	// Define a list of job inputs
	jobInputs := []string{"Job1", "Job2", "Job3"}

	var results []string
	for _, input := range jobInputs {
		result, err := jobBatcher.AddJob(input)
		assert.Nil(err)
		results = append(results, result.GetValue().(string))
	}
	// Ensure the results match the expected output
	assert.ElementsMatch([]string{"Processed - Job1", "Processed - Job2", "Processed - Job3"}, results)
}


func Test_WhenBatchSizeISLess(t *testing.T) {
	assert := assert.New(t)
	jobBatcher := NewBatcher(&prefixProcessor{Prefix: "Processed - "})
	jobBatcher.SetBatchSize(2)
	jobBatcher.SetFrequency(20*time.Millisecond)

	re1, err := jobBatcher.AddJob("1")
	assert.Nil(err)

	assert.Equal("Processed - 1", re1.GetValue())
}

func Test_WhenJobSizeIsMoreThanBatch(t *testing.T) {
	assert := assert.New(t)
	jobBatcher := NewBatcher(&prefixProcessor{Prefix: "Processed - "})
	jobBatcher.SetBatchSize(2)
	jobBatcher.SetFrequency(20*time.Millisecond)

	// Define a list of job inputs
	jobInputs := []string{"Job1", "Job2", "Job3","Job4","Job5"}
	var results []string
	for _, input := range jobInputs {
		result, err := jobBatcher.AddJob(input)
		assert.Nil(err)
		results = append(results, result.GetValue().(string))
	}

	assert.Equal([]string{"Processed - Job1", "Processed - Job2", "Processed - Job3", "Processed - Job4", "Processed - Job5"}, results)
}
// It tests that if the shutdown is called then it won't taken any other jobs and return the last processed
func Test_ReturnJobIfShutdown(t *testing.T) {
	assert := assert.New(t)

	jobBatcher := NewBatcher(&prefixProcessor{Prefix: "Processed - "})
	jobBatcher.SetBatchSize(2)
	jobBatcher.SetFrequency(20*time.Millisecond)

	re1, err := jobBatcher.AddJob("Job1")
	assert.Nil(err)

	jobBatcher.Stop()

	_, err = jobBatcher.AddJob("Job2")
	assert.Equal(ErrorMessage, err)

	assert.Equal("Processed - Job1", re1.GetValue())
}