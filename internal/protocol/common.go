package protocol

import (
	"github.com/gotrackery/protocol/common"
)

// Adapter is a generic adapter for the Position struct.
type Adapter interface {
	GenericPositions() []common.Position
}
