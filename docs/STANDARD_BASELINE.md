# Standard baseline

`go-bacnet` develops against **ANSI/ASHRAE 135-2024** with **Protocol Revision 31**
as the documented baseline. Conformance-oriented mapping references
**ANSI/ASHRAE 135.1-2023** descriptively (not as a certification claim).

Machine-readable edition / revision / tracked addenda:

[`standard-baseline.yaml`](standard-baseline.yaml)

ASHRAE has offered free read-only access to 135-2024 since August 2025
([bacnet.org/buy](https://bacnet.org/buy/)). The published addenda
([bacnet.org/addenda](https://bacnet.org/addenda/)), errata and
interpretations are the citable texts: each tracked item in the YAML names the
document it was checked against. The addenda list stays empty until an addendum
to 135-2024 itself is tracked.

The client **adapts** to remote capabilities (I-Am / Device object) and does not
require remote devices to implement Revision 31 features.

## Normative interpretation

- Document intentional interpretations and peer deviations in
  [COMPATIBILITY.md](COMPATIBILITY.md) and
  [bacnet-interop PEER_SUPPORT.md](https://github.com/otfabric/bacnet-interop/blob/main/PEER_SUPPORT.md).
- Do not weaken wire behaviour solely for a peer quirk; record the deviation.

## Where tests and fixtures live

| Kind | Location |
|---|---|
| Unit / codec tests | package tests under the module root and `service/`, `apdu/`, `npdu/`, `bvlc/`, … |
| Golden / codec fixtures | `bacnet-interop` `fixtures/codec/` via `BACNET_INTEROP_ROOT` or sibling checkout |
| Live peer tests | `go-bacnet/interop/` with `-tags=interop` |
| Capability declaration | [CLIENT_SUPPORT.md](CLIENT_SUPPORT.md) |
| Descriptive profile | [CLIENT_PROFILE.md](CLIENT_PROFILE.md) |
