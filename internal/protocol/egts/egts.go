package egts

import (
	"fmt"
	"strconv"

	"github.com/gotrackery/gotrackery/internal"
	"github.com/gotrackery/gotrackery/internal/tcp"
	"github.com/gotrackery/protocol/common"
	"github.com/gotrackery/protocol/egts"
)

var _ tcp.Protocol = (*EGTS)(nil)

const (
	Proto = "egts"
)

type EGTS struct {
}

func NewEGTS() *EGTS {
	return &EGTS{}
}

func (e *EGTS) Name() string {
	return Proto
}

func (e *EGTS) NewFrameSplitter() common.FrameSplitter {
	return egts.NewSplitter()
}

func (e *EGTS) Respond(s *internal.Session, bytes []byte) (res tcp.Result, err error) {
	pkg := egts.Packet{}
	_ = pkg.Decode(bytes)

	device := e.getDevice(&pkg)
	if device != "" {
		s.SetDevice(device)
	}

	res.Response, err = pkg.Response()
	if err != nil {
		return tcp.Result{CloseSession: true}, fmt.Errorf("got error on response: %s", err)
	}
	res.GenericAdapter = Adapter{Package: &pkg}
	return
}

func (e *EGTS) getDevice(pkg *egts.Packet) string {
	if pkg.ServicesFrameData == nil {
		return ""
	}
	if sfrd, ok := pkg.ServicesFrameData.(*egts.ServiceDataSet); ok {
		for _, record := range *sfrd {
			r := record
			if r.ObjectIDFieldExists == "1" {
				return strconv.FormatUint(uint64(r.ObjectIdentifier), 10)
			}
			for _, subRec := range r.RecordDataSet {
				if subRecData, ok := subRec.SubrecordData.(*egts.SrTermIdentity); ok {
					return strconv.FormatUint(uint64(subRecData.TerminalIdentifier), 10)
				}
			}
		}
	}
	return ""
}
