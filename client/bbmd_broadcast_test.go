// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/otfabric/go-bacnet"
	"github.com/otfabric/go-bacnet/bip"
	"github.com/otfabric/go-bacnet/bvlc"
	"github.com/otfabric/go-bacnet/internal/clock"
	"github.com/otfabric/go-bacnet/internal/diag"
	"github.com/otfabric/go-bacnet/internal/virtual"
)

// fdClient is a foreign-device client on a virtual link, its BBMD silent until
// the test accepts the registration.
func fdClient(t *testing.T) (*Client, *virtual.Transport, bip.Endpoint, clock.Clock) {
	t.Helper()
	clk := clock.NewManual(time.Unix(0, 0).UTC())
	local := bip.NewEndpoint(netip.MustParseAddrPort("10.0.0.5:47808"))
	bbmd := bip.NewEndpoint(netip.MustParseAddrPort("10.0.0.1:47808"))
	tr := virtual.New(local, clk, 16)
	c, err := New(
		WithTransport(AdaptVirtual(tr)),
		withClock(clk),
		WithForeignDevice(ForeignDeviceConfig{BBMD: bbmd, TTL: 60 * time.Second}),
		WithTransactionOptions(5*time.Second, 0, 0),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c, tr, bbmd, clk
}

// functions lists the BVLC functions of the frames the client sent.
func functions(tr *virtual.Transport) []bvlc.Function {
	var out []bvlc.Function
	for _, pkt := range tr.Outbox() {
		if msg, err := bvlc.Parse(pkt.Data, bacnet.DefaultDecodeLimits()); err == nil {
			out = append(out, msg.Function)
		}
	}
	return out
}

// accept answers the client's Register-Foreign-Device with a successful
// BVLC-Result.
func accept(t *testing.T, tr *virtual.Transport, bbmd bip.Endpoint, clk clock.Clock) {
	t.Helper()
	for deadline := time.Now().Add(time.Second); time.Now().Before(deadline); time.Sleep(5 * time.Millisecond) {
		for _, f := range functions(tr) {
			if f == bvlc.FunctionRegisterForeignDevice {
				result, err := bvlc.Append(nil, bvlc.Message{Function: bvlc.FunctionResult, ResultCode: 0})
				if err != nil {
					t.Fatal(err)
				}
				tr.Inject(virtual.InboundPacket{Data: result, ImmediatePeer: bbmd, ReceivedAt: clk.Now()})
				return
			}
		}
	}
	t.Fatal("timed out waiting for Register-Foreign-Device")
}

func TestForeignDeviceDoesNotBroadcastLocallyBeforeRegistration(t *testing.T) {
	c, tr, bbmd, clk := fdClient(t)
	_, err := c.Discover(context.Background(), DiscoveryOptions{Window: time.Millisecond})
	if !errors.Is(err, bacnet.ErrNotRegistered) {
		t.Fatalf("Discover before registration: %v, want ErrNotRegistered", err)
	}
	for _, f := range functions(tr) {
		if f == bvlc.FunctionOriginalBroadcastNPDU || f == bvlc.FunctionDistributeBroadcastToNetwork {
			t.Fatalf("a broadcast left before registration: %v", f)
		}
	}

	accept(t, tr, bbmd, clk)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := c.WaitForeignDeviceRegistered(ctx); err != nil {
		t.Fatal(err)
	}
	if err := c.SendWhoIs(context.Background(), c.broadcastEndpoint(), true, DiscoveryOptions{}); err != nil {
		t.Fatal(err)
	}
	var distributed bool
	for _, pkt := range tr.Outbox() {
		msg, err := bvlc.Parse(pkt.Data, bacnet.DefaultDecodeLimits())
		if err != nil {
			continue
		}
		if msg.Function == bvlc.FunctionOriginalBroadcastNPDU {
			t.Fatal("a registered foreign device broadcast locally")
		}
		if msg.Function == bvlc.FunctionDistributeBroadcastToNetwork && pkt.Destination == bbmd {
			distributed = true
		}
	}
	if !distributed {
		t.Fatal("no Distribute-Broadcast-To-Network to the BBMD after registration")
	}
}

func TestWaitForeignDeviceRegisteredEnds(t *testing.T) {
	c, _, _, _ := fdClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := c.WaitForeignDeviceRegistered(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("no answer from the BBMD: %v, want the context's deadline", err)
	}

	done := make(chan error, 1)
	go func() { done <- c.WaitForeignDeviceRegistered(context.Background()) }()
	_ = c.Close()
	select {
	case err := <-done:
		if !errors.Is(err, bacnet.ErrClosed) {
			t.Fatalf("closed while waiting: %v, want ErrClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close did not end the wait")
	}

	env := newVirtualPair(t)
	if err := env.Client.WaitForeignDeviceRegistered(context.Background()); !errors.Is(err, bacnet.ErrInvalidConfig) {
		t.Fatalf("not a foreign device: %v, want ErrInvalidConfig", err)
	}
}

// A registration that lapses closes the way again: the next waiter blocks until
// the BBMD accepts a new one.
func TestForeignDeviceReadyFollowsTheRegistration(t *testing.T) {
	bbmd := bip.NewEndpoint(netip.MustParseAddrPort("10.0.0.1:47808"))
	f, err := newFDState(ForeignDeviceConfig{BBMD: bbmd, TTL: time.Minute}, clock.Real{}, diag.Discard{})
	if err != nil {
		t.Fatal(err)
	}
	isOpen := func() bool {
		select {
		case <-f.readyCh():
			return true
		default:
			return false
		}
	}
	if isOpen() {
		t.Fatal("ready before any registration")
	}
	f.setRegistered(true)
	f.setRegistered(true) // a renewal: no second close
	if !isOpen() || !f.registered() {
		t.Fatal("not ready after an accepted registration")
	}
	f.setRegistered(false)
	f.setRegistered(false)
	if isOpen() || f.registered() {
		t.Fatal("still ready after the registration lapsed")
	}
	f.setRegistered(true)
	if !isOpen() {
		t.Fatal("not ready after registering again")
	}
}
