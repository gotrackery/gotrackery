package player

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	existingFolder = "test/path"
	existingFile1  = "test/path/file1.xml"
	existingFile2  = "test/path/file2.xml"
)

func TestProducer_Produce(t *testing.T) {
	filenames := make(chan string)
	quit := make(chan interface{})
	p := NewProducer(existingFolder, "*.xml", &filenames, quit)

	go p.Produce()

	// Test that the Produce method sends the correct filenames
	expected := []string{existingFile1, existingFile2}
	for _, expectedFilename := range expected {
		actualFilename := <-filenames
		assert.Equal(t, expectedFilename, actualFilename)
	}

	// Test that the Produce method closes the filenames channel
	_, ok := <-filenames
	assert.False(t, ok)
}

func TestProducer_find(t *testing.T) {
	p := NewProducer(existingFolder, "*.xml", nil, nil)

	// Test that the find method returns the correct filenames
	filenames, err := p.find(existingFolder, "*.xml")
	require.NoError(t, err)
	expected := []string{existingFile1, existingFile2}
	assert.Equal(t, expected, filenames)
}
