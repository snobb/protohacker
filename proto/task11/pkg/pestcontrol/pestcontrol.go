package pestcontrol

import (
	"context"
	"errors"
	"io"
	"log"
	"net"

	"proto/task11/pkg/authority"
	"proto/task11/pkg/frame"
)

// PestControl connection handler.
type PestControl struct {
	rw   io.ReadWriter
	addr net.Addr
}

// New creates a new pest control connection handler.
func New(rw io.ReadWriter, addr net.Addr) *PestControl {
	return &PestControl{rw: rw, addr: addr}
}

// Handle a single connection
func (p *PestControl) Handle(ctx context.Context) {
	if err := frame.Handshake(p.rw); err != nil {
		frame.WriteError(p.rw, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Print("cancelled...")
			return
		default:
		}

		frm, err := frame.ReadFrame(p.rw)
		if err != nil {
			frame.WriteError(p.rw, err)
			return
		}

		sv := frame.NewSiteVisit()
		if err := frm.UnloadInto(sv); err != nil {
			if !errors.Is(err, io.EOF) {
				log.Print("Handle:parse sitevisit:error:", err.Error())
				frame.WriteError(p.rw, err)
			}

			continue
		}

		if err := authority.HandleSite(ctx, sv); err != nil {
			log.Print("HandleSiteVisit error:", err.Error())
			frame.WriteError(p.rw, err)
		}
	}
}
