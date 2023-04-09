package egts

import (
	"fmt"
	"math"
	"strconv"

	gen "github.com/gotrackery/gotrackery/internal/protocol"
	"github.com/gotrackery/protocol/egts"
	"github.com/gotrackery/protocol/generic"
	"github.com/peterstace/simplefeatures/geom"
	"gopkg.in/guregu/null.v4"
)

var _ gen.Adapter = (*Adapter)(nil)

type Adapter struct {
	Package *egts.Package
}

func (a Adapter) GenericPositions() []generic.Position {
	var p generic.Position
	valid := a.Package.ErrorCode == egts.EgtsPcOk
	ps := make([]generic.Position, 0, 1)
	if a.Package.PacketType != egts.PtAppdataPacket ||
		a.Package.ServicesFrameData == nil {
		return nil
	}
	if sfrd, ok := a.Package.ServicesFrameData.(*egts.ServiceDataSet); ok {
		for _, record := range *sfrd {
			r := record
			p = a.convertRecordToGeneric(r, valid)
			if p.DeviceID != "" {
				ps = append(ps, p)
			}
		}
	}

	if len(ps) > 0 {
		return ps
	}
	return nil
}

func (a Adapter) convertRecordToGeneric(r egts.ServiceDataRecord, valid bool) generic.Position {
	p := generic.Position{
		Protocol: Proto,
	}

	var devFound bool
	if devFound = r.ObjectIDFieldExists == "1"; devFound {
		p.DeviceID = strconv.FormatUint(uint64(r.ObjectIdentifier), 10)
	}

	for _, subRec := range r.RecordDataSet {
		srec := subRec
		switch subRecData := srec.SubrecordData.(type) {
		case *egts.SrTermIdentity:
			p.DeviceID = strconv.FormatUint(uint64(subRecData.TerminalIdentifier), 10)
		case *egts.SrPosData:
			a.copyPosData(&p, subRecData, valid)
		case *egts.SrExtPosData:
			a.copyExtPosData(&p, subRecData)
		case *egts.SrAdSensorsData:
			a.copyAdSensorsData(&p, subRecData)
		case *egts.SrAbsAnSensData:
			// ToDo implement
		case *egts.SrAbsCntrData:
			// ToDo implement
		case *egts.SrLiquidLevelSensor:
			// ToDo implement
		}
	}

	return p
}

func (a Adapter) copyPosData(p *generic.Position, posData *egts.SrPosData, valid bool) {
	p.DeviceTime = posData.NavigationTime
	p.Location.X = math.Copysign(posData.Longitude, a.getHSSign(posData.LOHS))
	p.Location.Y = math.Copysign(posData.Latitude, a.getHSSign(posData.LAHS))
	p.Location.Type = geom.DimXY
	if posData.ALTE == "1" {
		p.Location.Type = geom.DimXYZ
		p.Location.Z = math.Copysign(float64(posData.Altitude), a.getAltSign(posData.AltitudeSign))
	}
	p.Location.Valid = valid && posData.VLD == "1"
	p.Speed = null.NewFloat(float64(posData.Speed), true)
	p.Course = null.NewFloat(float64(posData.Direction), true)
	p.Attributes = p.Attributes.AppendNullInt(
		generic.Odometer,
		null.NewInt(int64(posData.Odometer), true),
	)
	p.Attributes = p.Attributes.AppendNullString(
		generic.Move,
		null.NewString(posData.MV, true),
	)
	// ToDo add source data ???
}

func (a Adapter) copyExtPosData(p *generic.Position, extPosData *egts.SrExtPosData) {
	if extPosData.SatellitesFieldExists == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			generic.Satellites,
			null.NewInt(int64(extPosData.Satellites), true),
		)
	}
	if extPosData.PdopFieldExists == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			generic.PDOP,
			null.NewInt(int64(extPosData.PositionDilutionOfPrecision), true),
		)
	}
	if extPosData.HdopFieldExists == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			generic.HDOP,
			null.NewInt(int64(extPosData.HorizontalDilutionOfPrecision), true),
		)
	}
	if extPosData.VdopFieldExists == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			generic.VDOP,
			null.NewInt(int64(extPosData.VerticalDilutionOfPrecision), true),
		)
	}
	if extPosData.NavigationSystemFieldExists == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			generic.NavSystem,
			null.NewInt(int64(extPosData.NavigationSystem), true),
		)
	}
}

func (a Adapter) copyAdSensorsData(p *generic.Position, adSensorsData *egts.SrAdSensorsData) {
	/* ADIO - DigitalInputs */
	if adSensorsData.DigitalInputsOctetExists1 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_1", generic.DigInput),
			null.NewInt(int64(adSensorsData.AdditionalDigitalInputsOctet1), true),
		)
	}
	if adSensorsData.DigitalInputsOctetExists2 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_2", generic.DigInput),
			null.NewInt(int64(adSensorsData.AdditionalDigitalInputsOctet2), true),
		)
	}
	if adSensorsData.DigitalInputsOctetExists3 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_3", generic.DigInput),
			null.NewInt(int64(adSensorsData.AdditionalDigitalInputsOctet3), true),
		)
	}
	if adSensorsData.DigitalInputsOctetExists4 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_4", generic.DigInput),
			null.NewInt(int64(adSensorsData.AdditionalDigitalInputsOctet4), true),
		)
	}
	if adSensorsData.DigitalInputsOctetExists5 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_5", generic.DigInput),
			null.NewInt(int64(adSensorsData.AdditionalDigitalInputsOctet5), true),
		)
	}
	if adSensorsData.DigitalInputsOctetExists6 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_6", generic.DigInput),
			null.NewInt(int64(adSensorsData.AdditionalDigitalInputsOctet6), true),
		)
	}
	if adSensorsData.DigitalInputsOctetExists7 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_7", generic.DigInput),
			null.NewInt(int64(adSensorsData.AdditionalDigitalInputsOctet7), true),
		)
	}
	if adSensorsData.DigitalInputsOctetExists8 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_8", generic.DigInput),
			null.NewInt(int64(adSensorsData.AdditionalDigitalInputsOctet8), true),
		)
	}
	/* DOUT - DigitalOutputs */
	p.Attributes = p.Attributes.AppendNullInt(
		generic.DigOutput,
		null.NewInt(int64(adSensorsData.DigitalOutputs), true),
	)
	/* ANS - Analog sensors */
	if adSensorsData.AnalogSensorFieldExists1 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_1", generic.AnInput),
			null.NewInt(int64(adSensorsData.AnalogSensor1), true),
		)
	}
	if adSensorsData.AnalogSensorFieldExists2 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_2", generic.AnInput),
			null.NewInt(int64(adSensorsData.AnalogSensor2), true),
		)
	}
	if adSensorsData.AnalogSensorFieldExists3 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_3", generic.AnInput),
			null.NewInt(int64(adSensorsData.AnalogSensor3), true),
		)
	}
	if adSensorsData.AnalogSensorFieldExists4 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_4", generic.AnInput),
			null.NewInt(int64(adSensorsData.AnalogSensor4), true),
		)
	}
	if adSensorsData.AnalogSensorFieldExists5 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_5", generic.AnInput),
			null.NewInt(int64(adSensorsData.AnalogSensor5), true),
		)
	}
	if adSensorsData.AnalogSensorFieldExists6 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_6", generic.AnInput),
			null.NewInt(int64(adSensorsData.AnalogSensor6), true),
		)
	}
	if adSensorsData.AnalogSensorFieldExists7 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_7", generic.AnInput),
			null.NewInt(int64(adSensorsData.AnalogSensor7), true),
		)
	}
	if adSensorsData.AnalogSensorFieldExists8 == "1" {
		p.Attributes = p.Attributes.AppendNullInt(
			fmt.Sprintf("%s_8", generic.AnInput),
			null.NewInt(int64(adSensorsData.AnalogSensor8), true),
		)
	}
}

func (a Adapter) getHSSign(hemisphere string) float64 {
	if hemisphere == egts.LAHSNorth || hemisphere == egts.LOHSEast {
		return 1
	}
	return -1
}

func (a Adapter) getAltSign(s uint8) float64 {
	if s == egts.ALTSAboveSea {
		return 1
	}
	return -1
}
