package player

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

// Player is interface to replay original data from provided filename.
type Player interface {
	Play(filename string) error
}

// Consumer consumes a batch of files to process then concurrently.
type Consumer struct {
	filenames *chan string
	jobs      chan string
	replayer  Player
}

// NewConsumer creates instance of Consumer with given parameters.
func NewConsumer(l *chan string, j chan string, r Player) *Consumer {
	c := &Consumer{filenames: l, jobs: j, replayer: r}
	return c
}

// Work does work.
func (c *Consumer) Work(wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range c.jobs {
		if job == "" {
			return
		}
		err := c.replayer.Play(job)
		if err != nil {
			log.Error().Err(err).Str("job", job).Msg("replay failed")
			return
		}
	}
}

// Consume consumes workload and pushes for workers.
func (c *Consumer) Consume(ctx context.Context) {
	for {
		select {
		case job := <-*c.filenames:
			c.jobs <- job
		case <-ctx.Done():
			close(c.jobs)
			return
		}
	}
}
