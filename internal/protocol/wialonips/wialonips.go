package wialonips

import (
	"bufio"

	"github.com/gotrackery/gotrackery/internal"
	"github.com/gotrackery/gotrackery/internal/tcp"
	"github.com/gotrackery/protocol"
	"github.com/gotrackery/protocol/wialonips"
)

const (
	Proto      = "wialonips"
	ctxVersion = "version"
)

var _ tcp.Protocol = (*WialonIPS)(nil)

// GetSplitFunc returns the split function for the WialonIPS protocol.
func GetSplitFunc() bufio.SplitFunc {
	return wialonips.ScanPackage
}

// WialonIPS is a WialonIPS protocol struct.
type WialonIPS struct {
}

// NewWialonIPS creates a new WialonIPS struct instance.
func NewWialonIPS() *WialonIPS {
	return &WialonIPS{}
}

// GetName returns the name of the WialonIPS protocol.
func (w *WialonIPS) GetName() string {
	return Proto
}

// GetSplitFunc returns a split function for the WialonIPS protocol.
func (w *WialonIPS) GetSplitFunc() bufio.SplitFunc {
	return GetSplitFunc()
}

// Respond returns the result of parsing the WialonIPS data.
func (w *WialonIPS) Respond(s *internal.Session, bytes []byte) (res tcp.Result, err error) {
	pkg := wialonips.NewPackage(w.getVersion(s), s.GetDevice())
	err = pkg.Decode(bytes)
	if pkg.Type == wialonips.LoginPacket {
		s.SetDevice(pkg.IMEI)
		s.Set(ctxVersion, pkg.Version)
	}
	if pkg.Type == wialonips.UnknownPacket {
		return res, protocol.ErrInconsistentData
	}
	res.Response = pkg.Message.Response()
	if err != nil && (pkg.Type == wialonips.LoginPacket || pkg.Type == wialonips.UnknownPacket) {
		res.CloseSession = true
		return
	}

	res.GenericAdapter = Adapter{Package: &pkg}
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
