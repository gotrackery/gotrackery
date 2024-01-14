package event

import (
	"github.com/gotrackery/protocol/common"

	"github.com/gookit/event"
)

var _ event.Event = (*GenericEvent)(nil)

// GenericEvent wrapper for event.BasicEvent describes generic and unified events.
type GenericEvent struct {
	event.BasicEvent
}

// Position returns generic and unified position data.
func (e GenericEvent) Position() *common.Position {
	data := e.Data()
	if len(data) == 0 {
		return nil
	}
	pos, ok := data["position"]
	if !ok {
		return nil
	}
	position, ok := pos.(common.Position)
	if !ok {
		return nil
	}
	return &position
}

// SetPosition adds position data to an event.
func (e *GenericEvent) SetPosition(pos common.Position) {
	e.SetData(event.M{"position": pos})
}

type Reply struct {
	Error   error
	Message string
}

type Name string
const (
	PositionRecived Name = "position.received"
	CloseConnection Name = "close.connection"
)
