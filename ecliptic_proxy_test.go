package azimuth

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestPackSetProxies(t *testing.T) {
	addr := common.HexToAddress("0xF7306c5db0C1880FB2ed9c3972ad3e1A94999196")
	tests := []struct {
		name string
		sig  string
		pack func(uint32, common.Address) ([]byte, error)
	}{
		{name: "setManagementProxy", sig: "setManagementProxy(uint32,address)", pack: PackSetManagementProxy},
		{name: "setSpawnProxy", sig: "setSpawnProxy(uint16,address)", pack: PackSetSpawnProxy},
		{name: "setTransferProxy", sig: "setTransferProxy(uint32,address)", pack: PackSetTransferProxy},
		{name: "setVotingProxy", sig: "setVotingProxy(uint8,address)", pack: PackSetVotingProxy},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.pack(69, addr)
			if err != nil {
				t.Fatal(err)
			}
			sel := crypto.Keccak256([]byte(tc.sig))[:4]
			if !bytes.Equal(data[:4], sel) {
				t.Fatalf("selector %x want %x", data[:4], sel)
			}
			parsed, err := EclipticABI()
			if err != nil {
				t.Fatal(err)
			}
			unpacked, err := parsed.Methods[tc.name].Inputs.Unpack(data[4:])
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
		})
	}
}

func TestPackSetProxyRange(t *testing.T) {
	addr := common.HexToAddress("0xF7306c5db0C1880FB2ed9c3972ad3e1A94999196")
	if _, err := PackSetSpawnProxy(1566792653, addr); err == nil {
		t.Fatal("planet spawn proxy")
	}
	if _, err := PackSetVotingProxy(256, addr); err == nil {
		t.Fatal("star voting proxy")
	}
}

func TestBuyPlanetInnerGoldens(t *testing.T) {
	var fx struct {
		Prefix     uint32 `json:"prefix"`
		PrefixPatp string `json:"prefixPatp"`
		Point      uint32 `json:"point"`
		Inners     []struct {
			Method string `json:"method"`
			Input  string `json:"input"`
			Args   struct {
				Point   uint32 `json:"point"`
				Target  string `json:"target"`
				Manager string `json:"manager"`
				Reset   bool   `json:"reset"`
			} `json:"args"`
		} `json:"inners"`
	}
	loadJSON(t, "testdata/l1/buyPlanet-58132b57.json", &fx)
	if fx.Prefix != 44021 || fx.PrefixPatp != "~bossen" {
		t.Fatalf("prefix %d %s", fx.Prefix, fx.PrefixPatp)
	}
	pre, err := Prefix(fx.Point)
	if err != nil {
		t.Fatal(err)
	}
	if pre != fx.Prefix {
		t.Fatalf("computed prefix %d want %d", pre, fx.Prefix)
	}

	if len(fx.Inners) == 0 {
		t.Fatal("no inners")
	}
	for i, inner := range fx.Inners {
		var got []byte
		switch inner.Method {
		case "spawn":
			got, err = PackSpawn(inner.Args.Point, common.HexToAddress(inner.Args.Target))
		case "transferPoint":
			got, err = PackTransferPoint(inner.Args.Point, common.HexToAddress(inner.Args.Target), inner.Args.Reset)
		case "setManagementProxy":
			got, err = PackSetManagementProxy(inner.Args.Point, common.HexToAddress(inner.Args.Manager))
		default:
			t.Fatalf("inner %d: unknown method %s", i, inner.Method)
		}
		if err != nil {
			t.Fatalf("inner %d %s: %v", i, inner.Method, err)
		}
		want := mustHex(t, inner.Input)
		if !bytes.Equal(got, want) {
			t.Fatalf("inner %d %s: got %x want %x", i, inner.Method, got, want)
		}
	}
}
