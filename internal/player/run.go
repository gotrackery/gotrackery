package player

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Run starts replaying files from given path that satisfy mask with nConsumers goroutines.
func Run(path, mask string, replayer Player, nConsumers int) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	quit := make(chan interface{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		close(quit)
		cancelFunc()
	}()

	t := time.Now()
	log.Info().Msgf("Started in path %s", path)

	// runtime.GOMAXPROCS(runtime.NumCPU())
	files := make(chan string, 4) //nolint:gomnd

	p := NewProducer(path, mask, &files, quit)
	c := NewConsumer(&files, make(chan string, nConsumers), replayer)

	go p.Produce()
	go c.Consume(ctx)

	wg := &sync.WaitGroup{}
	wg.Add(nConsumers)
	for i := 0; i < nConsumers; i++ {
		go c.Work(wg)
	}

	wg.Wait()

	log.Info().Msgf("Spent %s", time.Since(t).String())
}
