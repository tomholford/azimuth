package azimuth

import "testing"

func TestParseFormatPoint(t *testing.T) {
	tests := []struct {
		name string
		n    uint32
	}{
		{name: "~zod", n: 0},
		{name: "~pet", n: 69},
		{name: "~marzod", n: 256},
		{name: "~fonlyd", n: 23501},
		{name: "~bossen", n: 44021},
		{name: "~fasten-ritbus", n: 1566792653},
		{name: "~libmer-lodnyt", n: 2552343541},
		{name: "~doprys-doppet", n: 2314841077},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParsePoint(tc.name)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.n {
				t.Fatalf("parse %s: %d want %d", tc.name, got, tc.n)
			}
			bare := tc.name[1:]
			gotBare, err := ParsePoint(bare)
			if err != nil {
				t.Fatal(err)
			}
			if gotBare != tc.n {
				t.Fatalf("parse %s: %d want %d", bare, gotBare, tc.n)
			}
			s, err := FormatPoint(tc.n)
			if err != nil {
				t.Fatal(err)
			}
			if s != tc.name {
				t.Fatalf("format %d: %s want %s", tc.n, s, tc.name)
			}
		})
	}
}

func TestParsePointErrors(t *testing.T) {
	tests := []struct {
		in string
	}{
		{in: ""},
		{in: "not-a-ship"},
		{in: "~nope"},
		{in: "~sampel-palnet-sampel-palnet"}, // moon
	}
	for _, tc := range tests {
		if _, err := ParsePoint(tc.in); err == nil {
			t.Fatalf("%q: expected error", tc.in)
		}
	}
}

func TestClan(t *testing.T) {
	tests := []struct {
		n    uint32
		want string
	}{
		{n: 0, want: "galaxy"},
		{n: 69, want: "galaxy"},
		{n: 256, want: "star"},
		{n: 23501, want: "star"},
		{n: 1566792653, want: "planet"},
	}
	for _, tc := range tests {
		got, err := Clan(tc.n)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("%d: %s want %s", tc.n, got, tc.want)
		}
	}
}

func TestPrefix(t *testing.T) {
	tests := []struct {
		n    uint32
		want uint32
	}{
		{n: 0, want: 0},
		{n: 69, want: 69},
		{n: 256, want: 0},
		{n: 23501, want: 205},
		{n: 44021, want: 245},
		{n: 1566792653, want: 23501},
		{n: 2552343541, want: 44021},
		{n: 2314841077, want: 44021},
	}
	for _, tc := range tests {
		got, err := Prefix(tc.n)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("%d: prefix %d want %d", tc.n, got, tc.want)
		}
	}
}
