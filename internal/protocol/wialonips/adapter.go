package wialonips

import (
	"fmt"

	gen "github.com/gotrackery/gotrackery/internal/protocol"
	"github.com/gotrackery/protocol"
	"github.com/gotrackery/protocol/generic"
	"github.com/gotrackery/protocol/wialonips"

	"golang.org/x/exp/maps"
	"gopkg.in/guregu/null.v4"
)

const (
	ibutton = "ibutton"
	inputs  = "inputs"
	outputs = "outputs"
	adc     = "adc"
)

var _ gen.Adapter = (*Adapter)(nil)

// Adapter is a generic adapter for the WialonIPS struct.
type Adapter struct {
	Package *wialonips.Package
}

// GenericPositions implements the generic.Adapter interface.
func (w Adapter) GenericPositions() (pos []generic.Position) {
	switch w.Package.Type {
	case wialonips.ShortenedDataPacket:
		m := w.Package.Message.(*wialonips.ShortenedDataMessage)
		p, err := w.convertShortenedToGeneric(m)
		if err != nil { // ToDo error handling
			return nil
		}
		return []generic.Position{p}
	case wialonips.DataPacket:
		m := w.Package.Message.(*wialonips.DataMessage)
		p, err := w.convertDataToGeneric(m)
		if err != nil { // ToDo error handling
			return nil
		}
		return []generic.Position{p}
	case wialonips.BlackBoxPacket:
		m := w.Package.Message.(*wialonips.BlackBoxMessage)
		pos = w.convertBlackBoxToGeneric(m)
		return
	}

	return nil
}

func (w Adapter) convertShortenedToGeneric(s *wialonips.ShortenedDataMessage) (generic.Position, error) {
	if s.Error() != nil {
		return generic.Position{}, s.Error()
	}
	p := generic.Position{
		Location:   s.Point.LocationXYZ(s.Altitude.Float64), // validness of Altitude dropped here
		Protocol:   Proto,
		DeviceID:   s.IMEI(),
		DeviceTime: s.RegisteredAt,
		Speed:      s.Speed,
		Course:     null.NewFloat(float64(s.Course.Int64), s.Course.Valid),
	}
	p.Attributes = p.Attributes.AppendNullInt(generic.Satellites, s.Sat)
	p.Attributes = p.Attributes.AppendNullString(generic.Proto, null.NewString(Proto, true))

	return p, nil
}

func (w Adapter) convertDataToGeneric(d *wialonips.DataMessage) (generic.Position, error) {
	if d.Error() != nil {
		return generic.Position{}, d.Error()
	}
	p, err := w.convertShortenedToGeneric(&d.ShortenedDataMessage)
	if err != nil {
		return generic.Position{}, fmt.Errorf("failed to convert data message to generic: %w", err)
	}
	p.DeviceID = d.IMEI()

	if p.Attributes == nil {
		p.Attributes = make(protocol.Attributes)
	}
	if d.Attributes != nil {
		maps.Copy(p.Attributes, d.Attributes)
	}
	p.Attributes = p.Attributes.AppendNullString(ibutton, d.IButton)
	p.Attributes = p.Attributes.AppendNullInt(inputs, d.Inputs)
	p.Attributes = p.Attributes.AppendNullInt(outputs, d.Outputs)
	p.Attributes = p.Attributes.AppendNullFloat(generic.HDOP, d.HDOP)
	p.Attributes = p.Attributes.AppendNullFloatSlice(adc, d.ADC)

	return p, nil
}

func (w Adapter) convertBlackBoxToGeneric(bb *wialonips.BlackBoxMessage) (pos []generic.Position) {
	if bb.Error() != nil {
		return nil
	}

	var (
		p   generic.Position
		err error
	)
	pos = make([]generic.Position, 0, len(bb.ShortenedMessages)+len(bb.DataMessages))
	for _, val := range bb.ShortenedMessages {
		p, err = w.convertShortenedToGeneric(&val)
		if err != nil {
			return nil
		}
		p.DeviceID = bb.IMEI()
		pos = append(pos, p)
	}
	for _, val := range bb.DataMessages {
		p, err = w.convertDataToGeneric(&val)
		if err != nil {
			return nil
		}
		p.DeviceID = bb.IMEI()
		pos = append(pos, p)
	}

	return
}
