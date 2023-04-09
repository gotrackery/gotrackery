package player

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// Producer producing work for parser from a list of xml files in given path.
type Producer struct {
	root, mask string
	filenames  *chan string
	quit       chan interface{}
}

// NewProducer creates of Producer instance.
func NewProducer(path, fileMask string, l *chan string, q chan interface{}) *Producer {
	return &Producer{root: path, mask: fileMask, filenames: l, quit: q}
}

// Produce sends filenames to consumer to parse files concurently.
func (p *Producer) Produce() {
	defer close(*p.filenames)
	fileList, err := p.find(p.root, p.mask)
	if err != nil {
		log.Error().Err(err).Msg("failed to find files")
		return
	}
	cnt := 0

loop:
	for _, filename := range fileList {
		select {
		case *p.filenames <- filename:
			cnt++
		case <-p.quit:
			break loop
		}
	}
	log.Info().Msgf("Processed %d files out of %d", cnt, len(fileList))
}

func (p *Producer) find(root, mask string) ([]string, error) {
	var a []string
	err := filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return fmt.Errorf("got error on walk dir: %w", e)
		}
		match, err := filepath.Match(mask, d.Name())
		if err != nil {
			return fmt.Errorf("failed to match file %s: %w", d.Name(), err)
		}
		if match {
			a = append(a, s)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get files from dir %s with mask %s: %w", root, mask, err)
	}
	return a, nil
}
