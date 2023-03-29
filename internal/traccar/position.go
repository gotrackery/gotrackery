package traccar

import (
	"encoding/json"
	"time"

	"github.com/gotrackery/protocol/generic"

	"gopkg.in/guregu/null.v4"
)

type Position struct {
	Protocol   string
	DeviceID   string
	ServerTime time.Time
	DeviceTime time.Time
	FixTime    time.Time
	Valid      bool
	Latitude   float64
	Longitude  float64
	Altitude   float64
	Speed      float64
	Course     float64
	Address    null.String
	Attributes string
	Accuracy   float64
	Network    null.String
}

func CreateFromGeneric(p generic.Position) (*Position, error) {
	attr, err := json.Marshal(p.Attributes)
	if err != nil {
		return nil, err
	}
	return &Position{
		Protocol:   p.Protocol,
		DeviceID:   p.DeviceID,
		ServerTime: time.Now(),
		DeviceTime: p.DeviceTime,
		FixTime:    p.DeviceTime,
		Valid:      p.Valid,
		Latitude:   p.Y,
		Longitude:  p.X,
		Altitude:   p.Z,
		Speed:      p.Speed.Float64,
		Course:     p.Course.Float64,
		Address:    null.NewString("", false),
		Attributes: string(attr),
		Accuracy:   0,
		Network:    null.NewString("", false),
	}, nil
}
