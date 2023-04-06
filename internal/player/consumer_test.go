package player

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsumer_Work(t *testing.T) {
	// Setup
	wg := &sync.WaitGroup{}
	jobs := make(chan string)
	filenames := make(chan string)
	replayer := &mockPlayer{}
	consumer := NewConsumer(&filenames, jobs, replayer)

	// Exercise
	wg.Add(1)
	go consumer.Work(wg)
	jobs <- "filename"
	close(jobs)
	wg.Wait()

	// Verify
	assert.Equal(t, "filename", replayer.filename)
}

func TestConsumer_Consume(t *testing.T) {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan string)
	filenames := make(chan string)
	replayer := &mockPlayer{}
	consumer := NewConsumer(&filenames, jobs, replayer)

	// Exercise
	go consumer.Consume(ctx)
	filenames <- "filename"
	cancel()

	// Verify
	assert.Equal(t, "filename", <-jobs)
	assert.Equal(t, 0, len(jobs))
}

type mockPlayer struct {
	mux      sync.Mutex
	filename string
}

func (m *mockPlayer) Play(filename string) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.filename = filename
	return nil
}
