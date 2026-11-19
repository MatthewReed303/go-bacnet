// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"net/netip"
	"sync"
	"testing"

	"github.com/otfabric/go-bacnet"
	"github.com/otfabric/go-bacnet/apdu"
	"github.com/otfabric/go-bacnet/bip"
	"github.com/otfabric/go-bacnet/bvlc"
	"github.com/otfabric/go-bacnet/npdu"
	"github.com/otfabric/go-bacnet/server/object"
	"github.com/otfabric/go-bacnet/service"
)

func TestRequesterOf(t *testing.T) {
	inBubble(t, testRequesterOf)
}

func testRequesterOf(t *testing.T) {
	if _, ok := RequesterOf(context.Background()); ok {
		t.Fatal("a context of no request has a requester")
	}
	var mu sync.Mutex
	var got []Requester
	var oks []bool
	record := func(ctx context.Context) {
		r, ok := RequesterOf(ctx)
		mu.Lock()
		got, oks = append(got, r), append(oks, ok)
		mu.Unlock()
	}
	routerEP := bip.NewEndpoint(netip.MustParseAddrPort("10.0.0.9:47808"))
	groups := make(chan struct{}, 1)
	h := newHarness(t, WithHooks(Hooks{
		BeforeWrite: func(ctx context.Context, _ bacnet.ObjectIdentifier, _ object.Write) error { record(ctx); return nil },
		WriteGroup:  func(ctx context.Context, _ service.WriteGroup) { record(ctx); groups <- struct{}{} },
	}))
	enc := mustEncode(t)
	write := enc(service.EncodeWriteProperty(service.WritePropertyRequest{Object: av1, Property: pvRef, Value: bacnet.RealValue(2)}))

	// A local station: the datagram's sender, no NPDU source.
	h.confirmed(1, apdu.ServiceWriteProperty, write)
	h.expectType(apdu.TypeSimpleACK)
	// A station on network 5 behind the router at routerEP.
	routed := apdu.AppendConfirmedRequest(nil, apdu.ConfirmedRequest{MaxSegments: 0, MaxAPDU: 5, InvokeID: 2, ServiceChoice: apdu.ServiceWriteProperty, Payload: write})
	h.injectFrom(routerEP, bvlc.FunctionOriginalUnicastNPDU, npdu.NPDU{Version: npdu.Version1, ExpectingReply: true,
		Source: bacnet.RemoteStation(5, bacnet.MustMAC([]byte{7})), APDU: routed})
	h.expectType(apdu.TypeSimpleACK)
	// An unconfirmed request reaches its hook with its requester too.
	h.unconfirmed(apdu.ServiceWriteGroup, enc(service.EncodeWriteGroup(service.WriteGroup{GroupNumber: 1, WritePriority: 8,
		ChangeList: []bacnet.Element{{Context: true, TagNumber: 0, Value: bacnet.UnsignedValue(1)}}})))
	<-groups

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 3 || !oks[0] || !oks[1] || !oks[2] {
		t.Fatalf("requesters %+v %v", got, oks)
	}
	if !got[0].Endpoint.Equal(clientEP) || !got[0].Via.Equal(clientEP) || got[0].Source != (bacnet.Address{}) {
		t.Fatalf("local requester %+v", got[0])
	}
	if !got[1].Endpoint.Equal(routerEP) || !got[1].Via.Equal(routerEP) || !got[1].Source.Equal(bacnet.RemoteStation(5, bacnet.MustMAC([]byte{7}))) {
		t.Fatalf("routed requester %+v", got[1])
	}
	if !got[2].Endpoint.Equal(clientEP) {
		t.Fatalf("unconfirmed requester %+v", got[2])
	}
}
