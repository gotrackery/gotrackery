package wialonips

import (
	"fmt"

	gen "github.com/gotrackery/gotrackery/internal/protocol"
	"github.com/gotrackery/protocol/common"
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

// Adapter is a common adapter for the WialonIPS struct.
type Adapter struct {
	Packet *wialonips.Packet
}

// GenericPositions implements the common.Adapter interface.
func (w Adapter) GenericPositions() (pos []common.Position) {
	switch w.Packet.Type {
	case wialonips.ShortenedDataPacket:
		m := w.Packet.Message.(*wialonips.ShortenedDataMessage)
		p, err := w.convertShortenedTocommon(m)
		if err != nil { // ToDo error handling
			return nil
		}
		return []common.Position{p}
	case wialonips.DataPacket:
		m := w.Packet.Message.(*wialonips.DataMessage)
		p, err := w.convertDataTocommon(m)
		if err != nil { // ToDo error handling
			return nil
		}
		return []common.Position{p}
	case wialonips.BlackBoxPacket:
		m := w.Packet.Message.(*wialonips.BlackBoxMessage)
		pos = w.convertBlackBoxTocommon(m)
		return
	}

	return nil
}

func (w Adapter) convertShortenedTocommon(s *wialonips.ShortenedDataMessage) (common.Position, error) {
	if s.Error() != nil {
		return common.Position{}, s.Error()
	}
	p := common.Position{
		Location:   s.Point.LocationXYZ(s.Altitude.Float64), // validness of Altitude dropped here
		Protocol:   Proto,
		DeviceID:   s.IMEI(),
		DeviceTime: s.RegisteredAt,
		Speed:      s.Speed,
		Course:     null.NewFloat(float64(s.Course.Int64), s.Course.Valid),
	}
	p.Attributes = p.Attributes.AppendNullInt(common.Satellites, s.Sat)
	p.Attributes = p.Attributes.AppendNullString(common.Proto, null.NewString(Proto, true))

	return p, nil
}

func (w Adapter) convertDataTocommon(d *wialonips.DataMessage) (common.Position, error) {
	if d.Error() != nil {
		return common.Position{}, d.Error()
	}
	p, err := w.convertShortenedTocommon(&d.ShortenedDataMessage)
	if err != nil {
		return common.Position{}, fmt.Errorf("failed to convert data message to common: %w", err)
	}
	p.DeviceID = d.IMEI()

	if p.Attributes == nil {
		p.Attributes = make(common.Attributes)
	}
	if d.Attributes != nil {
		maps.Copy(p.Attributes, d.Attributes)
	}
	p.Attributes = p.Attributes.AppendNullString(ibutton, d.IButton)
	p.Attributes = p.Attributes.AppendNullInt(inputs, d.Inputs)
	p.Attributes = p.Attributes.AppendNullInt(outputs, d.Outputs)
	p.Attributes = p.Attributes.AppendNullFloat(common.HDOP, d.HDOP)
	p.Attributes = p.Attributes.AppendNullFloatSlice(adc, d.ADC)

	return p, nil
}

func (w Adapter) convertBlackBoxTocommon(bb *wialonips.BlackBoxMessage) (pos []common.Position) {
	if bb.Error() != nil {
		return nil
	}

	var (
		p   common.Position
		err error
	)
	pos = make([]common.Position, 0, len(bb.ShortenedMessages)+len(bb.DataMessages))
	for _, val := range bb.ShortenedMessages {
		p, err = w.convertShortenedTocommon(&val)
		if err != nil {
			return nil
		}
		p.DeviceID = bb.IMEI()
		pos = append(pos, p)
	}
	for _, val := range bb.DataMessages {
		p, err = w.convertDataTocommon(&val)
		if err != nil {
			return nil
		}
		p.DeviceID = bb.IMEI()
		pos = append(pos, p)
	}

	return
}
