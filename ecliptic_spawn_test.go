package azimuth

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestPackSpawnGolden(t *testing.T) {
	var fx struct {
		Input string `json:"input"`
		Args  struct {
			Point  uint32 `json:"point"`
			Target string `json:"target"`
		} `json:"args"`
	}
	loadJSON(t, "testdata/l1/spawn-05319969.json", &fx)
	got, err := PackSpawn(fx.Args.Point, common.HexToAddress(fx.Args.Target))
	if err != nil {
		t.Fatal(err)
	}
	want := mustHex(t, fx.Input)
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x want %x", got, want)
	}
}

func TestPackEscapeGolden(t *testing.T) {
	var fx struct {
		Input string `json:"input"`
		Args  struct {
			Point       uint32 `json:"point"`
			Patp        string `json:"patp"`
			Sponsor     uint32 `json:"sponsor"`
			SponsorPatp string `json:"sponsorPatp"`
		} `json:"args"`
	}
	loadJSON(t, "testdata/l1/escape-b6920b85.json", &fx)
	if fx.Args.Patp != "~dosbyl" || fx.Args.SponsorPatp != "~rus" {
		t.Fatalf("names %s %s", fx.Args.Patp, fx.Args.SponsorPatp)
	}
	got, err := PackEscape(fx.Args.Point, fx.Args.Sponsor)
	if err != nil {
		t.Fatal(err)
	}
	want := mustHex(t, fx.Input)
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x want %x", got, want)
	}
}

func TestPackTransferPointGolden(t *testing.T) {
	var fx struct {
		Input string `json:"input"`
		Args  struct {
			Point      uint32 `json:"point"`
			Patp       string `json:"patp"`
			Prefix     uint32 `json:"prefix"`
			PrefixPatp string `json:"prefixPatp"`
			Target     string `json:"target"`
			Reset      bool   `json:"reset"`
		} `json:"args"`
	}
	loadJSON(t, "testdata/l1/transferPoint-88b933f2.json", &fx)
	if fx.Args.Patp != "~tabdyl" || fx.Args.PrefixPatp != "~dyl" {
		t.Fatalf("names %s %s", fx.Args.Patp, fx.Args.PrefixPatp)
	}
	pre, err := Prefix(fx.Args.Point)
	if err != nil {
		t.Fatal(err)
	}
	if pre != fx.Args.Prefix {
		t.Fatalf("prefix %d want %d", pre, fx.Args.Prefix)
	}
	got, err := PackTransferPoint(fx.Args.Point, common.HexToAddress(fx.Args.Target), fx.Args.Reset)
	if err != nil {
		t.Fatal(err)
	}
	want := mustHex(t, fx.Input)
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x want %x", got, want)
	}
}

func TestPackTransferPointSelector(t *testing.T) {
	addr := common.HexToAddress("0xF7306c5db0C1880FB2ed9c3972ad3e1A94999196")
	data, err := PackTransferPoint(69, addr, true)
	if err != nil {
		t.Fatal(err)
	}
	sel := crypto.Keccak256([]byte("transferPoint(uint32,address,bool)"))[:4]
	if !bytes.Equal(data[:4], sel) {
		t.Fatalf("selector %x want %x", data[:4], sel)
	}
	parsed, err := EclipticABI()
	if err != nil {
		t.Fatal(err)
	}
	unpacked, err := parsed.Methods["transferPoint"].Inputs.Unpack(data[4:])
	if err != nil {
		t.Fatal(err)
	}
	n, err := unpackUint32(unpacked[0])
	if err != nil {
		t.Fatal(err)
	}
	if n != 69 {
		t.Fatalf("point %v", unpacked[0])
	}
	got, err := unpackAddress(unpacked[1])
	if err != nil {
		t.Fatal(err)
	}
	if got != addr {
		t.Fatalf("addr %s want %s", got.Hex(), addr.Hex())
	}
	reset, ok := unpacked[2].(bool)
	if !ok || !reset {
		t.Fatalf("reset %v", unpacked[2])
	}
}

func TestPackEscapeFamily(t *testing.T) {
	tests := []struct {
		name string
		sig  string
		pack func() ([]byte, error)
		arg0 uint32
		arg1 uint32
		narg int
	}{
		{
			name: "escape",
			sig:  "escape(uint32,uint32)",
			pack: func() ([]byte, error) { return PackEscape(69, 256) },
			arg0: 69,
			arg1: 256,
			narg: 2,
		},
		{
			name: "cancelEscape",
			sig:  "cancelEscape(uint32)",
			pack: func() ([]byte, error) { return PackCancelEscape(69) },
			arg0: 69,
			narg: 1,
		},
		{
			name: "adopt",
			sig:  "adopt(uint32)",
			pack: func() ([]byte, error) { return PackAdopt(69) },
			arg0: 69,
			narg: 1,
		},
		{
			name: "reject",
			sig:  "reject(uint32)",
			pack: func() ([]byte, error) { return PackReject(69) },
			arg0: 69,
			narg: 1,
		},
		{
			name: "detach",
			sig:  "detach(uint32)",
			pack: func() ([]byte, error) { return PackDetach(69) },
			arg0: 69,
			narg: 1,
		},
	}
	parsed, err := EclipticABI()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.pack()
			if err != nil {
				t.Fatal(err)
			}
			sel := crypto.Keccak256([]byte(tc.sig))[:4]
			if !bytes.Equal(data[:4], sel) {
				t.Fatalf("selector %x want %x", data[:4], sel)
			}
			unpacked, err := parsed.Methods[tc.name].Inputs.Unpack(data[4:])
			if err != nil {
				t.Fatal(err)
			}
			if len(unpacked) != tc.narg {
				t.Fatalf("nargs %d want %d", len(unpacked), tc.narg)
			}
			n, err := unpackUint32(unpacked[0])
			if err != nil {
				t.Fatal(err)
			}
			if n != tc.arg0 {
				t.Fatalf("arg0 %v want %d", unpacked[0], tc.arg0)
			}
			if tc.narg == 2 {
				n, err = unpackUint32(unpacked[1])
				if err != nil {
					t.Fatal(err)
				}
				if n != tc.arg1 {
					t.Fatalf("arg1 %v want %d", unpacked[1], tc.arg1)
				}
			}
		})
	}
}
