// SPDX-License-Identifier: MIT

package bacnet

import (
	"regexp"
	"testing"
)

var standardName = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// Every standard name is spelled as clause 21 writes it, names one value, and
// reads back to it; the constants' names are untouched by the other table.
func TestStandardNamesRoundTrip(t *testing.T) {
	seen := map[string]bool{}
	for _, m := range []map[ObjectType]string{objectTypeNames, standardObjectTypeNames} {
		for v, name := range m {
			if !standardName.MatchString(name) || seen[name] {
				t.Fatalf("object type %d: %q is not a standard name, or names two types", v, name)
			}
			seen[name] = true
			if got, ok := ParseObjectType(name); !ok || got != v || v.String() != name {
				t.Fatalf("%q: parsed %d (%v), String %q, want %d", name, got, ok, v.String(), v)
			}
		}
	}
	for v := range standardObjectTypeNames {
		if _, dup := objectTypeNames[v]; dup {
			t.Fatalf("object type %d is named twice", v)
		}
	}
	seen = map[string]bool{}
	for _, m := range []map[PropertyIdentifier]string{propertyIdentifierNames, standardPropertyIdentifierNames} {
		for v, name := range m {
			if !standardName.MatchString(name) || seen[name] {
				t.Fatalf("property %d: %q is not a standard name, or names two properties", v, name)
			}
			seen[name] = true
			if got, ok := ParsePropertyIdentifier(name); !ok || got != v || v.String() != name {
				t.Fatalf("%q: parsed %d (%v), String %q, want %d", name, got, ok, v.String(), v)
			}
		}
	}
	for v := range standardPropertyIdentifierNames {
		if _, dup := propertyIdentifierNames[v]; dup {
			t.Fatalf("property %d is named twice", v)
		}
	}
}

func TestStandardNamesSpotChecks(t *testing.T) {
	for v, want := range map[ObjectType]string{
		12: "loop", 16: "program", 41: "datepattern-value", 43: "datetimepattern-value", 49: "timepattern-value",
		53: "channel", 54: "lighting-output", 59: "lift",
	} {
		if got := v.String(); got != want {
			t.Fatalf("ObjectType(%d) = %q, want %q", uint16(v), got, want)
		}
	}
	for v, want := range map[PropertyIdentifier]string{
		336: "supported-security-algorithms", 438: "bacnet-ipv6-udp-port", 417: "command",
	} {
		if got := v.String(); got != want {
			t.Fatalf("PropertyIdentifier(%d) = %q, want %q", uint32(v), got, want)
		}
	}
	if ObjectType(1000).String() != "ObjectType(1000)" || PropertyIdentifier(4000).String() != "PropertyIdentifier(4000)" {
		t.Fatal("an unnamed value no longer formats as its number")
	}
	if _, ok := ParseObjectType("analog_input"); ok {
		t.Fatal("a name not in clause 21 form was taken")
	}
	if _, ok := ParsePropertyIdentifier(""); ok {
		t.Fatal("an empty name was taken")
	}
	if p, ok := ParsePropertyIdentifier("present-value"); !ok || p != PropertyPresentValue {
		t.Fatalf("present-value = %d, %v", p, ok)
	}
}
