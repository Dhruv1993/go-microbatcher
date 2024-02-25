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
