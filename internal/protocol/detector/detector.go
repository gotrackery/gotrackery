package detector

import (
	"bufio"
	"bytes"

	"github.com/gotrackery/gotrackery/internal"
	"github.com/gotrackery/gotrackery/internal/protocol/egts"
	"github.com/gotrackery/gotrackery/internal/protocol/wialonips"
	"github.com/gotrackery/gotrackery/internal/tcp/server"
	egts2 "github.com/gotrackery/protocol/egts"
	ips "github.com/gotrackery/protocol/wialonips"
)

var _ server.Protocol = (*Detector)(nil)

const (
	Proto        = "detector"
	unknownProto = "unknown"
	bce          = "bce"
)

func GetSplitFunc() bufio.SplitFunc {
	scanner := ProtocolScanner{}
	return scanner.ScanProtocol
}

type Detector struct {
	detectedProto string
	dummyResponse []byte
}

func NewDetector(resp []byte) *Detector {
	return &Detector{detectedProto: "unknown", dummyResponse: resp}
}

func (d *Detector) GetName() string {
	return Proto
}

func (d *Detector) GetSplitFunc() bufio.SplitFunc {
	return GetSplitFunc()
}

func (d *Detector) Respond(s *internal.Session, bytes []byte) (res server.Result, err error) {
	d.detectedProto = protoDetector(bytes)
	s.SetDevice(d.detectedProto)
	res.CloseSession = false // let it close by timeout
	res.Response = d.dummyResponse
	res.GenericAdapter = nil
	return
}

type ProtocolScanner struct {
	detectedProto string
}

func NewProtocolScanner() *ProtocolScanner {
	return &ProtocolScanner{detectedProto: unknownProto}
}

func (ps *ProtocolScanner) ScanProtocol(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	p := protoDetector(data)
	switch {
	// case ps.detectedProto == bce:
	// 	return 0, nil, nil
	case p == egts.Proto:
		ps.detectedProto = egts.Proto
		return egts2.ScanPackage(data, atEOF)
	// case p == bce:
	// 	ps.detectedProto = bce
	// 	return 0, nil, nil
	case p == wialonips.Proto:
		ps.detectedProto = wialonips.Proto
		return ips.ScanPackage(data, atEOF)
	}

	// If we're at EOF, return it.
	if atEOF {
		return len(data), data, nil
	}
	// return every 16 bytes
	if len(data) < 16 {
		return 0, nil, nil
	}
	return 16, data[:16], nil
}

func protoDetector(data []byte) string {
	switch {
	case data[0] == 0x01:
		return egts.Proto
	case len(data) > 7 && string(data[:7]) == "#BCE#\r\n":
		return bce
	case len(data) > 7:
		bytesSet := bytes.Split(data, []byte("#"))
		if len(bytesSet) == 3 {
			return wialonips.Proto
		}
	}
	return unknownProto
}
