// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"testing"
	"time"

	"github.com/otfabric/go-bacnet"
	"github.com/otfabric/go-bacnet/apdu"
	"github.com/otfabric/go-bacnet/npdu"
	"github.com/otfabric/go-bacnet/service"
)

func whoAmIRequest(t *testing.T) []byte {
	t.Helper()
	payload, err := service.EncodeWhoAmI(service.WhoAmI{VendorID: 999, ModelName: "VAV-2", SerialNumber: "SN-0042"})
	if err != nil {
		t.Fatal(err)
	}
	return apdu.AppendUnconfirmedRequest(nil, apdu.UnconfirmedRequest{ServiceChoice: apdu.ServiceWhoAmI, Payload: payload})
}

func waitWhoAmI(t *testing.T, got <-chan WhoAmIDelivery) WhoAmIDelivery {
	t.Helper()
	select {
	case d := <-got:
		return d
	case <-time.After(time.Second):
		t.Fatal("no Who-Am-I delivered")
	}
	return WhoAmIDelivery{}
}

func TestWhoAmIReachesTheHandler(t *testing.T) {
	got := make(chan WhoAmIDelivery, 4)
	env := newVirtualPair(t, WithWhoAmIHandler(func(d WhoAmIDelivery) { got <- d }))

	// Broadcast on the local network: the sender's B/IP address.
	injectBroadcastNPDU(t, env.ClientTr, env.Peer, time.Now(), whoAmIRequest(t))
	d := waitWhoAmI(t, got)
	if d.Request.VendorID != 999 || d.Request.ModelName != "VAV-2" || d.Request.SerialNumber != "SN-0042" {
		t.Fatalf("request %+v", d.Request)
	}
	if !d.Address.Equal(env.Target.Address) || d.Immediate != env.Peer {
		t.Fatalf("from %v via %v, want %v via %v", d.Address, d.Immediate, env.Target.Address, env.Peer)
	}

	// From an MS/TP station behind a router: its network and station address, the router as the immediate peer.
	station := bacnet.RemoteStation(5, bacnet.MustMAC([]byte{12}))
	injectUnicastNPDUWithSource(t, env.ClientTr, env.Peer, time.Now(), npdu.NPDU{Version: npdu.Version1, Source: station, APDU: whoAmIRequest(t)})
	d = waitWhoAmI(t, got)
	if !d.Address.Equal(station) || d.Immediate != env.Peer {
		t.Fatalf("routed Who-Am-I from %v via %v", d.Address, d.Immediate)
	}

	// A You-Are answers it at that address.
	before := len(env.ClientTr.Outbox())
	payload, err := service.EncodeYouAre(service.YouAre{VendorID: 999, ModelName: "VAV-2", SerialNumber: "SN-0042",
		Device: bacnet.ObjectIdentifier{Type: bacnet.ObjectTypeDevice, Instance: 2042}})
	if err != nil {
		t.Fatal(err)
	}
	if err := env.Client.InvokeUnconfirmed(context.Background(), Target{Address: d.Address, Endpoint: d.Immediate}, apdu.ServiceYouAre, payload); err != nil {
		t.Fatal(err)
	}
	if n := sentNPDU(t, env, before); !n.Destination.Equal(station) {
		t.Fatalf("You-Are sent to %v, want %v", n.Destination, station)
	}
}

func TestWhoAmIMalformedOrAPanickingHandler(t *testing.T) {
	got := make(chan WhoAmIDelivery, 4)
	calls := 0
	env := newVirtualPair(t, WithWhoAmIHandler(func(d WhoAmIDelivery) {
		calls++ // the receive loop calls the handler from one goroutine
		if calls == 1 {
			panic("handler fault")
		}
		got <- d
	}))
	bad := apdu.AppendUnconfirmedRequest(nil, apdu.UnconfirmedRequest{ServiceChoice: apdu.ServiceWhoAmI, Payload: []byte{0x21}})
	injectBroadcastNPDU(t, env.ClientTr, env.Peer, time.Now(), bad)              // malformed: reported, not delivered
	injectBroadcastNPDU(t, env.ClientTr, env.Peer, time.Now(), whoAmIRequest(t)) // the handler panics: recovered
	injectBroadcastNPDU(t, env.ClientTr, env.Peer, time.Now(), whoAmIRequest(t))
	if d := waitWhoAmI(t, got); d.Request.SerialNumber != "SN-0042" {
		t.Fatalf("after a malformed request and a panic: %+v", d)
	}
	select {
	case d := <-got:
		t.Fatalf("one more delivery than requests handled: %+v", d)
	case <-time.After(50 * time.Millisecond):
	}
	// Without a handler a Who-Am-I is ignored.
	env.Client.SetWhoAmIHandler(nil)
	injectBroadcastNPDU(t, env.ClientTr, env.Peer, time.Now(), whoAmIRequest(t))
	select {
	case d := <-got:
		t.Fatalf("delivered with no handler: %+v", d)
	case <-time.After(50 * time.Millisecond):
	}
}
