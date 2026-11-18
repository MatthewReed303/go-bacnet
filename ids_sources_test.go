// SPDX-License-Identifier: MIT

package bacnet

import "testing"

// TestStandardNamesSources pins every name this module takes from an addendum to the
// ASHRAE document that assigned it, as its BACnetObjectType or
// BACnetPropertyIdentifier ASN.1 writes it. The documents are free from
// bacnet.org/addenda; third-party tables differ on some of these spellings.
func TestStandardNamesSources(t *testing.T) {
	objects := []struct {
		source string
		value  ObjectType
		name   string
	}{
		{"135-2008w", 41, "date-pattern-value"},
		{"135-2008w", 43, "datetime-pattern-value"},
		{"135-2008w", 49, "time-pattern-value"},
		{"135-2016bd", 60, "staging"},
		{"135-2016bi", 62, "audit-reporter"},
		{"135-2020ca", 63, "color"},
		{"135-2020ca", 64, "color-temperature"},
	}
	for _, c := range objects {
		if got := c.value.String(); got != c.name {
			t.Errorf("%s: ObjectType(%d) = %q, want %q", c.source, uint16(c.value), got, c.name)
		}
		if got, ok := ParseObjectType(c.name); !ok || got != c.value {
			t.Errorf("%s: ParseObjectType(%q) = %d, %v", c.source, c.name, got, ok)
		}
	}
	properties := []struct {
		source string
		value  PropertyIdentifier
		name   string
	}{
		{"135-2016bd", 492, "default-present-value"},
		{"135-2016bd", 493, "present-stage"},
		{"135-2016bd", 494, "stages"},
		{"135-2016bd", 495, "stage-names"},
		{"135-2016bd", 496, "target-references"},
		{"135-2016bi", 497, "audit-source-reporter"},
		{"135-2016bi", 498, "audit-level"},
		{"135-2016bi", 499, "audit-notification-recipient"},
		{"135-2016bi", 500, "audit-priority-filter"},
		{"135-2016bi", 501, "auditable-operations"},
		{"135-2016bi", 502, "delete-on-forward"},
		{"135-2016bi", 503, "maximum-send-delay"},
		{"135-2016bi", 504, "monitored-objects"},
		{"135-2016bi", 505, "send-now"},
		{"135-2016bt", 506, "floor-number"},
		{"135-2016bj", 507, "device-uuid"},
		{"135-2020cc", 508, "additional-reference-ports"},
		{"135-2020cc", 509, "certificate-signing-request-file"},
		{"135-2020cc", 510, "command-validation-result"},
		{"135-2020cc", 511, "issuer-certificate-files"},
		{"135-2020cc", 4194304, "max-bvlc-length-accepted"},
		{"135-2020cc", 4194305, "max-npdu-length-accepted"},
		{"135-2020cc", 4194306, "operational-certificate-file"},
		{"135-2020cc", 4194307, "current-health"},
		{"135-2020cc", 4194308, "sc-connect-wait-timeout"},
		{"135-2020cc", 4194309, "sc-direct-connect-accept-enable"},
		{"135-2020cc", 4194310, "sc-direct-connect-accept-uris"},
		{"135-2020cc", 4194311, "sc-direct-connect-binding"},
		{"135-2020cc", 4194312, "sc-direct-connect-connection-status"},
		{"135-2020cc", 4194313, "sc-direct-connect-initiate-enable"},
		{"135-2020cc", 4194314, "sc-disconnect-wait-timeout"},
		{"135-2020cc", 4194315, "sc-failed-connection-requests"},
		{"135-2020cc", 4194316, "sc-failover-hub-connection-status"},
		{"135-2020cc", 4194317, "sc-failover-hub-uri"},
		{"135-2020cc", 4194318, "sc-hub-connector-state"},
		{"135-2020cc", 4194319, "sc-hub-function-accept-uris"},
		{"135-2020cc", 4194320, "sc-hub-function-binding"},
		{"135-2020cc", 4194321, "sc-hub-function-connection-status"},
		{"135-2020cc", 4194322, "sc-hub-function-enable"},
		{"135-2020cc", 4194323, "sc-heartbeat-timeout"},
		{"135-2020cc", 4194324, "sc-primary-hub-connection-status"},
		{"135-2020cc", 4194325, "sc-primary-hub-uri"},
		{"135-2020cc", 4194326, "sc-maximum-reconnect-time"},
		{"135-2020cc", 4194327, "sc-minimum-reconnect-time"},
		{"135-2020ca", 4194328, "color-override"},
		{"135-2020ca", 4194329, "color-reference"},
		{"135-2020ca", 4194330, "default-color"},
		{"135-2020ca", 4194331, "default-color-temperature"},
		{"135-2020ca", 4194332, "override-color-reference"},
		{"135-2020bv", 4194333, "write-every-scheduled-action"},
		{"errata to 135-2020ca (2023)", 4194334, "color-command"},
		{"errata to 135-2020ca (2023)", 4194335, "high-end-trim"},
		{"errata to 135-2020ca (2023)", 4194336, "low-end-trim"},
		{"errata to 135-2020ca (2023)", 4194337, "trim-fade-time"},
		{"135-2020bx", 4194338, "device-address-proxy-enable"},
		{"135-2020bx", 4194339, "device-address-proxy-table"},
		{"135-2020bx", 4194340, "device-address-proxy-timeout"},
		{"135-2020cj", 4194341, "default-on-value"},
		{"135-2020cj", 4194342, "last-on-value"},
		{"135-2020cp", 4194343, "authorization-cache"},
		{"135-2020cp", 4194344, "authorization-groups"},
		{"135-2020cp", 4194345, "authorization-policy"},
		{"135-2020cp", 4194346, "authorization-scope"},
		{"135-2020cp", 4194347, "authorization-server"},
		{"135-2020cp", 4194348, "authorization-status"},
		{"135-2020cp", 4194349, "max-proxied-i-ams-per-second"},
	}
	for _, c := range properties {
		if got := c.value.String(); got != c.name {
			t.Errorf("%s: PropertyIdentifier(%d) = %q, want %q", c.source, uint32(c.value), got, c.name)
		}
		if got, ok := ParsePropertyIdentifier(c.name); !ok || got != c.value {
			t.Errorf("%s: ParsePropertyIdentifier(%q) = %d, %v", c.source, c.name, got, ok)
		}
	}
	// 135-2024 is the baseline: 135-2024cu assigns 4194350 and later.
	if s := PropertyIdentifier(4194350).String(); s != "PropertyIdentifier(4194350)" {
		t.Errorf("PropertyIdentifier(4194350) = %q, a value assigned after 135-2024", s)
	}
}

// The joined spellings of the pattern-value types parse, and String keeps
// the clause 21 form.
func TestObjectTypeAliases(t *testing.T) {
	for name, want := range map[string]ObjectType{
		"datepattern-value": 41, "datetimepattern-value": 43, "timepattern-value": 49,
	} {
		got, ok := ParseObjectType(name)
		if !ok || got != want {
			t.Fatalf("ParseObjectType(%q) = %d, %v, want %d", name, got, ok, want)
		}
		if got.String() == name {
			t.Fatalf("String returned the alias %q", name)
		}
	}
}
