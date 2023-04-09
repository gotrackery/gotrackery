package player

import (
	"testing"
)

func TestRun(t *testing.T) {
	// Test with valid parameters
	path := "test/path"
	mask := "*.xml"
	replayer := &mockPlayer{}
	nConsumers := 2

	Run(path, mask, replayer, nConsumers)

	// Test with invalid parameters
	path = ""
	mask = ""
	replayer = &mockPlayer{}
	nConsumers = 0

	Run(path, mask, replayer, nConsumers)
}
