package wialonips

import (
	"github.com/gotrackery/gotrackery/internal"
	"github.com/gotrackery/gotrackery/internal/tcp"
	"github.com/gotrackery/protocol/common"
	"github.com/gotrackery/protocol/wialonips"
)

const (
	Proto      = "wialonips"
	ctxVersion = "version"
)

var _ tcp.Protocol = (*WialonIPS)(nil)

// WialonIPS is a WialonIPS protocol struct.
type WialonIPS struct {
}

// NewWialonIPS creates a new WialonIPS struct instance.
func NewWialonIPS() *WialonIPS {
	return &WialonIPS{}
}

// Name returns the name of the WialonIPS protocol.
func (w *WialonIPS) Name() string {
	return Proto
}

// NewFrameSplitter returns a new instance of the split function for the WialonIPS protocol.
func (w *WialonIPS) NewFrameSplitter() common.FrameSplitter {
	return wialonips.NewSplitter()
}

// Respond returns the result of parsing the WialonIPS data.
func (w *WialonIPS) Respond(s *internal.Session, bytes []byte) (res tcp.Result, err error) {
	pkg := wialonips.NewPacket(w.getVersion(s), s.GetDevice())
	err = pkg.Decode(bytes)
	if pkg.Type == wialonips.LoginPacket {
		s.SetDevice(pkg.IMEI)
		s.Set(ctxVersion, pkg.Version)
	}
	if pkg.Type == wialonips.UnknownPacket {
		return res, common.ErrBadData
	}
	res.Response = pkg.Message.Response()
	if err != nil && (pkg.Type == wialonips.LoginPacket || pkg.Type == wialonips.UnknownPacket) {
		res.CloseSession = true
		return
	}

	res.GenericAdapter = Adapter{Packet: &pkg}
	return
}

func (w *WialonIPS) getVersion(s *internal.Session) wialonips.Version {
	val := s.Get(ctxVersion)
	if val == nil {
		return wialonips.UnknownVersion
	}
	if ver, ok := val.(wialonips.Version); ok {
		return ver
	}
	return wialonips.UnknownVersion
}
