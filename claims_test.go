package azimuth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestPackAddClaim(t *testing.T) {
	data, err := PackAddClaim(69, "prot1", "claim", []byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	sel := crypto.Keccak256([]byte("addClaim(uint32,string,string,bytes)"))[:4]
	if !bytes.Equal(data[:4], sel) {
		t.Fatalf("selector %x want %x", data[:4], sel)
	}
	parsed, err := ClaimsABI()
	if err != nil {
		t.Fatal(err)
	}
	unpacked, err := parsed.Methods["addClaim"].Inputs.Unpack(data[4:])
	if err != nil {
		t.Fatal(err)
	}
	if unpacked[0].(uint32) != 69 {
		t.Fatalf("point %v", unpacked[0])
	}
	if unpacked[1].(string) != "prot1" || unpacked[2].(string) != "claim" {
		t.Fatalf("fields %v %v", unpacked[1], unpacked[2])
	}
	if !bytes.Equal(unpacked[3].([]byte), []byte{0x01}) {
		t.Fatalf("dossier %x", unpacked[3])
	}
}

func TestPackAddClaimEmptyDossier(t *testing.T) {
	data, err := PackAddClaim(0, "prot1", "claim", nil)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ClaimsABI()
	if err != nil {
		t.Fatal(err)
	}
	unpacked, err := parsed.Methods["addClaim"].Inputs.Unpack(data[4:])
	if err != nil {
		t.Fatal(err)
	}
	if len(unpacked[3].([]byte)) != 0 {
		t.Fatalf("dossier %x", unpacked[3])
	}
}

func TestPackAddClaimEmptyFields(t *testing.T) {
	if _, err := PackAddClaim(0, "", "claim", nil); err == nil {
		t.Fatal("empty protocol")
	}
	if _, err := PackAddClaim(0, "prot1", "", nil); err == nil {
		t.Fatal("empty claim")
	}
}

func TestPackRemoveClaim(t *testing.T) {
	data, err := PackRemoveClaim(69, "prot1", "claim")
	if err != nil {
		t.Fatal(err)
	}
	sel := crypto.Keccak256([]byte("removeClaim(uint32,string,string)"))[:4]
	if !bytes.Equal(data[:4], sel) {
		t.Fatalf("selector %x want %x", data[:4], sel)
	}
	parsed, err := ClaimsABI()
	if err != nil {
		t.Fatal(err)
	}
	unpacked, err := parsed.Methods["removeClaim"].Inputs.Unpack(data[4:])
	if err != nil {
		t.Fatal(err)
	}
	if unpacked[0].(uint32) != 69 || unpacked[1].(string) != "prot1" || unpacked[2].(string) != "claim" {
		t.Fatalf("%v", unpacked)
	}
}

func TestPackRemoveClaimEmptyFields(t *testing.T) {
	if _, err := PackRemoveClaim(0, "prot1", ""); err == nil {
		t.Fatal("empty claim")
	}
}

func TestPackClearClaims(t *testing.T) {
	data, err := PackClearClaims(69)
	if err != nil {
		t.Fatal(err)
	}
	sel := crypto.Keccak256([]byte("clearClaims(uint32)"))[:4]
	if !bytes.Equal(data[:4], sel) {
		t.Fatalf("selector %x want %x", data[:4], sel)
	}
	parsed, err := ClaimsABI()
	if err != nil {
		t.Fatal(err)
	}
	unpacked, err := parsed.Methods["clearClaims"].Inputs.Unpack(data[4:])
	if err != nil {
		t.Fatal(err)
	}
	if unpacked[0].(uint32) != 69 {
		t.Fatalf("point %v", unpacked[0])
	}
}

func TestGetClaimMockRPC(t *testing.T) {
	c, err := NewClaimsContract()
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.ABI.Methods["claims"].Outputs.Pack("prot1", "claim", []byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	srv := mockEthRPC(t, map[string]any{
		"eth_call": hexutil.Encode(out),
	})
	defer srv.Close()

	client, err := ethclient.Dial(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	got, err := GetClaim(context.Background(), client, 69, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got.Protocol != "prot1" || got.Claim != "claim" || !bytes.Equal(got.Dossier, []byte{0x01}) {
		t.Fatalf("%+v", got)
	}
}

func TestGetClaimIndexRange(t *testing.T) {
	if _, err := GetClaim(context.Background(), nil, 0, MaxClaims); err == nil {
		t.Fatal("expected error")
	}
}

func TestFindClaimMockRPC(t *testing.T) {
	c, err := NewClaimsContract()
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.ABI.Methods["findClaim"].Outputs.Pack(uint8(3))
	if err != nil {
		t.Fatal(err)
	}
	srv := mockEthRPC(t, map[string]any{
		"eth_call": hexutil.Encode(out),
	})
	defer srv.Close()

	client, err := ethclient.Dial(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	got, err := FindClaim(context.Background(), client, 69, "prot1", "claim")
	if err != nil {
		t.Fatal(err)
	}
	if got != 3 {
		t.Fatalf("index %d", got)
	}
}

func TestFindClaimEmptyFields(t *testing.T) {
	if _, err := FindClaim(context.Background(), nil, 0, "", "claim"); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetClaimsMockRPC(t *testing.T) {
	c, err := NewClaimsContract()
	if err != nil {
		t.Fatal(err)
	}
	filled, err := c.ABI.Methods["claims"].Outputs.Pack("prot1", "claim", []byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	empty, err := c.ABI.Methods["claims"].Outputs.Pack("", "", []byte{})
	if err != nil {
		t.Fatal(err)
	}
	claimsSel := hexutil.Encode(mustPack(t, c, "claims", uint32(69), big.NewInt(0))[:4])

	srv := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		switch method {
		case "eth_chainId", "net_version":
			return "0x1", nil
		case "eth_call":
			data := ethCallData(params)
			if !strings.HasPrefix(data, claimsSel) {
				return nil, fmt.Errorf("unexpected call %s", data)
			}
			idx := claimIndexFromCall(t, c, data)
			if idx == 0 {
				return hexutil.Encode(filled), nil
			}
			return hexutil.Encode(empty), nil
		default:
			return nil, fmt.Errorf("unexpected %s", method)
		}
	})
	defer srv.Close()

	client, err := ethclient.Dial(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	got, err := GetClaims(context.Background(), client, 69)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Protocol != "prot1" || got[0].Claim != "claim" {
		t.Fatalf("%+v", got)
	}
}

func claimIndexFromCall(t *testing.T, c *Contract, data string) int64 {
	t.Helper()
	raw := mustHex(t, data)
	if len(raw) < 4 {
		t.Fatalf("calldata %s", data)
	}
	unpacked, err := c.ABI.Methods["claims"].Inputs.Unpack(raw[4:])
	if err != nil {
		t.Fatal(err)
	}
	switch n := unpacked[1].(type) {
	case *big.Int:
		return n.Int64()
	default:
		t.Fatalf("index type %T", unpacked[1])
		return 0
	}
}

func TestUnpackClaimRoundTrip(t *testing.T) {
	c, err := NewClaimsContract()
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.ABI.Methods["claims"].Outputs.Pack("twitter", "@zod", []byte("proof"))
	if err != nil {
		t.Fatal(err)
	}
	unpacked, err := c.ABI.Unpack("claims", out)
	if err != nil {
		t.Fatal(err)
	}
	got, err := unpackClaim(unpacked)
	if err != nil {
		t.Fatal(err)
	}
	if got.Protocol != "twitter" || got.Claim != "@zod" || !bytes.Equal(got.Dossier, []byte("proof")) {
		t.Fatalf("%+v", got)
	}
}
