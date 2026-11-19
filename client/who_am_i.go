// SPDX-License-Identifier: MIT

package client

import (
	"fmt"

	"github.com/otfabric/go-bacnet"
	"github.com/otfabric/go-bacnet/bip"
	"github.com/otfabric/go-bacnet/internal/diag"
	"github.com/otfabric/go-bacnet/service"
)

// WhoAmIDelivery is a Who-Am-I a device sent (clause 16.11): its vendor,
// model name and serial number, and where it came from. A You-Are for it goes
// to that address (InvokeUnconfirmed with Address and Immediate as the
// Target), or as a broadcast.
type WhoAmIDelivery struct {
	Request   service.WhoAmI
	Address   bacnet.Address
	Origin    bip.Endpoint
	Immediate bip.Endpoint
}

// WhoAmIHandler receives decoded Who-Am-I requests.
//
// The handler runs synchronously on the receive path: it must not block and
// must not call Client methods that wait for a response. Handler panics are
// recovered so a faulty callback cannot stop reception.
type WhoAmIHandler func(WhoAmIDelivery)

// WithWhoAmIHandler registers a handler for inbound Who-Am-I requests, the
// requests a device without a configured identity sends to be given one with
// You-Are. Nil disables the handler; without one a Who-Am-I is ignored.
func WithWhoAmIHandler(h WhoAmIHandler) Option {
	return func(c *config) { c.whoAmIHandler = h }
}

// SetWhoAmIHandler installs or clears the inbound Who-Am-I handler. It may be
// called concurrently with reception.
func (c *Client) SetWhoAmIHandler(h WhoAmIHandler) {
	c.eventMu.Lock()
	c.whoAmIHandler = h
	c.eventMu.Unlock()
}

func (c *Client) deliverWhoAmI(req service.WhoAmI, src packetSource) {
	c.eventMu.Lock()
	h := c.whoAmIHandler
	c.eventMu.Unlock()
	if h == nil {
		return
	}
	d := WhoAmIDelivery{Request: req, Address: src.bacnetAddress, Origin: src.origin, Immediate: src.immediate}
	if d.Address.MAC().IsZero() {
		if addr, ok := bipMACAddress(src.origin); ok {
			d.Address = addr
		}
	}
	defer func() {
		if rec := recover(); rec != nil {
			c.diag.Report(diag.Event{
				Kind:    diag.KindCallback,
				Message: "Who-Am-I handler panic",
				Fields:  map[string]any{"panic": fmt.Sprint(rec)},
			})
		}
	}()
	h(d)
}
