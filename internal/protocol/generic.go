package protocol

import (
	"github.com/gotrackery/protocol/generic"
)

// Adapter is a generic adapter for the Position struct.
type Adapter interface {
	GenericPositions() []generic.Position
}
