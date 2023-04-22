package tcp

import (
	"github.com/gotrackery/gotrackery/internal"
	"github.com/gotrackery/protocol/common"
)

// Protocol is the tcp protocol contract.
type Protocol interface {
	// Name returns the name of the protocol.
	Name() string
	// NewFrameSplitter returns instance of the split function for the protocol.
	NewFrameSplitter() common.FrameSplitter
	// Respond returns the Result as passed was processed.
	Respond(*internal.Session, []byte) (Result, error)
}
