package sampler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_Accept(t *testing.T) {

	useCases := []struct {
		description string
		goalPCT     float64
		testCount   int
	}{

		{
			description: "50% goal",
			goalPCT:     50,
			testCount:   100000,
		},

		{
			description: "95.1% goal",
			goalPCT:     95.1,
			testCount:   100000,
		},
	}

	for _, useCase := range useCases {
		sampler := New(useCase.goalPCT)
		acceptCount := 0
		for i := 0; i < useCase.testCount; i++ {
			if sampler.Accept() {
				acceptCount++
			}
		}

		actualAcceptPCT := int(100.0 * (float64(acceptCount) / float64(useCase.testCount)))
		if actualAcceptPCT > int(useCase.goalPCT) { //allows -1 diff
			actualAcceptPCT--
		}
		if actualAcceptPCT < int(useCase.goalPCT) { //allows +1 diff
			actualAcceptPCT++
		}
		assert.Equal(t, int(useCase.goalPCT), actualAcceptPCT, useCase.description)
	}
}

func TestService_AcceptWithThreshold(t *testing.T) {
	sampler := New(100) // should not affect the result
	acceptCount := 0
	testcount := 100000
	for i := 0; i < testcount; i++ {
		if sampler.AcceptWithThreshold(50.0) {
			acceptCount++
		}
	}
	actualAcceptPCT := int(100.0 * (float64(acceptCount) / float64(testcount)))
	if actualAcceptPCT > int(50.0) { //allows -1 diff
		actualAcceptPCT--
	}
	if actualAcceptPCT < int(50.0) { //allows +1 diff
		actualAcceptPCT++
	}
	assert.Equal(t, int(50), actualAcceptPCT, "accept with threshold off ")
}
