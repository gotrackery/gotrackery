package event

import (
	"github.com/gotrackery/protocol/common"

	"github.com/gookit/event"
)

var _ event.Event = (*GenericEvent)(nil)

// GenericEvent describes generic and unified events extracted from received data.
type GenericEvent struct {
	event.BasicEvent
	common.Position
}

// GetPosition returns generic and unified position data.
func (e GenericEvent) GetPosition() common.Position {
	return e.Position
}
