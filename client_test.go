package azimuth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestDialEmpty(t *testing.T) {
	if _, err := Dial("", ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestClientGetPointRoller(t *testing.T) {
	roller := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		if method != "getPoint" {
			return nil, fmt.Errorf("unexpected %s", method)
		}
		res := Result{Dominion: DominionL2}
		res.Network.Keys.Life = "1"
		res.Network.Keys.Suite = "1"
		res.Network.Keys.Auth = "0x" + strings.Repeat("11", 32)
		res.Network.Keys.Crypt = "0x" + strings.Repeat("22", 32)
		return res, nil
	})
	defer roller.Close()
	eth := mockEthRPC(t, nil)
	defer eth.Close()

	c, err := Dial(eth.URL, roller.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	p, err := c.GetPoint(context.Background(), 69)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "~pet" || p.Dominion != DominionL2 || p.Revision != 1 {
		t.Fatalf("%+v", p)
	}
}

func TestClientGetPointL1Fallback(t *testing.T) {
	az, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	var crypt, auth [32]byte
	copy(crypt[:], bytes.Repeat([]byte{0xaa}, 32))
	copy(auth[:], bytes.Repeat([]byte{0xbb}, 32))
	owner := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	mgmt := common.HexToAddress("0xF7306c5db0C1880FB2ed9c3972ad3e1A94999196")
	spawn := common.HexToAddress("0xd1428F18A8255C0984291CEBFc83E6F982F7De9f")
	vote := common.HexToAddress("0x3333333333333333333333333333333333333333")
	xfer := common.HexToAddress("0x4444444444444444444444444444444444444444")

	pointsOut, err := az.ABI.Methods["points"].Outputs.Pack(
		crypt, auth, true, true, true, uint32(256), uint32(512), uint32(1), uint32(3), uint32(7),
	)
	if err != nil {
		t.Fatal(err)
	}
	rightsOut, err := az.ABI.Methods["rights"].Outputs.Pack(owner, mgmt, spawn, vote, xfer)
	if err != nil {
		t.Fatal(err)
	}
	p := getPointL1Fallback(t, az, pointsOut, rightsOut)
	if p.Dominion != DominionL1 || p.Owner.Address != owner || p.Revision != 3 || p.Rift != 7 {
		t.Fatalf("%+v", p)
	}
	if p.ManagementProxy.Address != mgmt || p.SpawnProxy.Address != spawn || p.VotingProxy.Address != vote || p.TransferProxy.Address != xfer {
		t.Fatalf("proxies %+v", p)
	}
	if !p.HasSponsor || p.Sponsor != 256 || !p.Active || !p.EscapeRequested || p.EscapeRequestedTo != 512 {
		t.Fatalf("network %+v", p)
	}
	if !bytes.Equal(p.CryptKey, crypt[:]) {
		t.Fatalf("crypt %x", p.CryptKey)
	}
}

func TestClientGetPointL1FallbackDeposited(t *testing.T) {
	az, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	var z [32]byte
	pointsOut, err := az.ABI.Methods["points"].Outputs.Pack(
		z, z, false, true, false, uint32(0), uint32(0), uint32(0), uint32(0), uint32(0),
	)
	if err != nil {
		t.Fatal(err)
	}
	zero := common.Address{}
	rightsOut, err := az.ABI.Methods["rights"].Outputs.Pack(DepositAddr(), zero, zero, zero, zero)
	if err != nil {
		t.Fatal(err)
	}
	p := getPointL1Fallback(t, az, pointsOut, rightsOut)
	if p.Dominion != DominionL2 || p.Owner.Address != DepositAddr() {
		t.Fatalf("%+v", p)
	}
}

func TestClientGetPointL1FallbackSpawn(t *testing.T) {
	az, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	var z [32]byte
	owner := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	pointsOut, err := az.ABI.Methods["points"].Outputs.Pack(
		z, z, true, true, false, uint32(0), uint32(0), uint32(1), uint32(1), uint32(0),
	)
	if err != nil {
		t.Fatal(err)
	}
	zero := common.Address{}
	rightsOut, err := az.ABI.Methods["rights"].Outputs.Pack(owner, zero, DepositAddr(), zero, zero)
	if err != nil {
		t.Fatal(err)
	}
	p := getPointL1Fallback(t, az, pointsOut, rightsOut)
	if p.Dominion != DominionSpawn || p.Owner.Address != owner || p.SpawnProxy.Address != DepositAddr() {
		t.Fatalf("%+v", p)
	}
}

func TestClientConfigureKeysL1(t *testing.T) {
	crypt := bytes.Repeat([]byte{0xaa}, 32)
	auth := bytes.Repeat([]byte{0xbb}, 32)
	from := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")

	roller := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		if method != "getPoint" {
			return nil, fmt.Errorf("unexpected %s", method)
		}
		return Result{Dominion: DominionL1}, nil
	})
	defer roller.Close()
	eth := mockEthRPC(t, nil)
	defer eth.Close()

	c, err := Dial(eth.URL, roller.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	u, err := c.ConfigureKeys(context.Background(), 69, crypt, auth, CryptoSuiteVersion, false, from)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL1 || u.To != EclipticAddr() {
		t.Fatalf("%+v", u)
	}
	want, err := PackConfigureKeys(69, crypt, auth, CryptoSuiteVersion, false)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(u.Data, want) {
		t.Fatalf("data %x want %x", u.Data, want)
	}
}

func TestClientConfigureKeysL2(t *testing.T) {
	crypt := bytes.Repeat([]byte{0x11}, 32)
	auth := bytes.Repeat([]byte{0x22}, 32)
	from := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")

	roller := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		switch method {
		case "getPoint":
			res := Result{Dominion: DominionL2}
			res.Ownership.Owner.Address = from.Hex()
			res.Ownership.Owner.Nonce = 3
			return res, nil
		case "prepareForSigning":
			return "0xcafe", nil
		default:
			return nil, fmt.Errorf("unexpected %s", method)
		}
	})
	defer roller.Close()
	eth := mockEthRPC(t, nil)
	defer eth.Close()

	c, err := Dial(eth.URL, roller.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	u, err := c.ConfigureKeys(context.Background(), 69, crypt, auth, CryptoSuiteVersion, true, from)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL2 || u.Hash != "0xcafe" || u.Tx != L2TxConfigureKeys || u.Nonce != 3 {
		t.Fatalf("%+v", u)
	}
	if u.From.Ship != "~pet" || u.From.Proxy != ProxyOwn {
		t.Fatalf("from %+v", u.From)
	}
}

func TestClientSetManagementProxyL1(t *testing.T) {
	mgr := common.HexToAddress("0xF7306c5db0C1880FB2ed9c3972ad3e1A94999196")
	from := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	roller := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		if method != "getPoint" {
			return nil, fmt.Errorf("unexpected %s", method)
		}
		return Result{Dominion: DominionL1}, nil
	})
	defer roller.Close()
	eth := mockEthRPC(t, nil)
	defer eth.Close()

	c, err := Dial(eth.URL, roller.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	u, err := c.SetManagementProxy(context.Background(), 69, mgr, from)
	if err != nil {
		t.Fatal(err)
	}
	want, err := PackSetManagementProxy(69, mgr)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL1 || !bytes.Equal(u.Data, want) {
		t.Fatalf("%+v", u)
	}
}

func TestClientSetSpawnProxySpawnDominion(t *testing.T) {
	sp := common.HexToAddress("0xF7306c5db0C1880FB2ed9c3972ad3e1A94999196")
	from := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	var sawTx string
	roller := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		switch method {
		case "getPoint":
			res := Result{Dominion: DominionSpawn}
			res.Ownership.Owner.Address = from.Hex()
			res.Ownership.Owner.Nonce = 1
			return res, nil
		case "prepareForSigning":
			var p struct {
				Tx string `json:"tx"`
			}
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, err
			}
			sawTx = p.Tx
			return "0xddd", nil
		default:
			return nil, fmt.Errorf("unexpected %s", method)
		}
	})
	defer roller.Close()
	eth := mockEthRPC(t, nil)
	defer eth.Close()

	c, err := Dial(eth.URL, roller.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	u, err := c.SetSpawnProxy(context.Background(), 256, sp, from)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL2 || u.Hash != "0xddd" || sawTx != L2TxSetSpawnProxy {
		t.Fatalf("layer=%s hash=%s tx=%s", u.Layer, u.Hash, sawTx)
	}
}

func TestClientSpawnL1(t *testing.T) {
	target := common.HexToAddress("0xfbBe7c8B6009bD5cee28A7d7A68231A4104C3d85")
	from := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		if method != "getPoint" {
			return nil, fmt.Errorf("unexpected %s", method)
		}
		return Result{Dominion: DominionL1}, nil
	})

	u, err := c.Spawn(context.Background(), 1566792653, target, from)
	if err != nil {
		t.Fatal(err)
	}
	want, err := PackSpawn(1566792653, target)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL1 || !bytes.Equal(u.Data, want) {
		t.Fatalf("%+v", u)
	}
}

func TestClientSpawnSpawnDominion(t *testing.T) {
	target := common.HexToAddress("0xfbBe7c8B6009bD5cee28A7d7A68231A4104C3d85")
	from := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	var saw struct {
		Tx   string    `json:"tx"`
		From L2From    `json:"from"`
		Data SpawnData `json:"data"`
	}
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		switch method {
		case "getPoint":
			res := Result{Dominion: DominionSpawn}
			res.Ownership.Owner.Address = from.Hex()
			res.Ownership.Owner.Nonce = 4
			return res, nil
		case "prepareForSigning":
			if err := json.Unmarshal(params, &saw); err != nil {
				return nil, err
			}
			return "0xsp", nil
		default:
			return nil, fmt.Errorf("unexpected %s", method)
		}
	})

	u, err := c.Spawn(context.Background(), 1566792653, target, from)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL2 || u.Hash != "0xsp" || u.Tx != L2TxSpawn {
		t.Fatalf("%+v", u)
	}
	if saw.From.Ship != "~fonlyd" || saw.From.Proxy != ProxyOwn {
		t.Fatalf("from %+v", saw.From)
	}
	if saw.Data.Ship != "~fasten-ritbus" || !strings.EqualFold(saw.Data.Address, target.Hex()) {
		t.Fatalf("data %+v", saw.Data)
	}
}

func TestClientTransferPointL1(t *testing.T) {
	target := common.HexToAddress("0xd1428F18A8255C0984291CEBFc83E6F982F7De9f")
	from := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		if method != "getPoint" {
			return nil, fmt.Errorf("unexpected %s", method)
		}
		return Result{Dominion: DominionL1}, nil
	})

	u, err := c.TransferPoint(context.Background(), 69, target, true, from)
	if err != nil {
		t.Fatal(err)
	}
	want, err := PackTransferPoint(69, target, true)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL1 || !bytes.Equal(u.Data, want) {
		t.Fatalf("%+v", u)
	}
}

func TestClientDepositL1(t *testing.T) {
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		if method != "getPoint" {
			return nil, fmt.Errorf("unexpected %s", method)
		}
		return testPointResult(DominionL1), nil
	})
	u, err := c.Deposit(context.Background(), 256)
	if err != nil {
		t.Fatal(err)
	}
	want, err := PackDeposit(256)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL1 || u.To != EclipticAddr() || !bytes.Equal(u.Data, want) {
		t.Fatalf("%+v", u)
	}
}

func TestClientDepositSpawn(t *testing.T) {
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		if method != "getPoint" {
			return nil, fmt.Errorf("unexpected %s", method)
		}
		return testPointResult(DominionSpawn), nil
	})
	u, err := c.Deposit(context.Background(), 256)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL1 {
		t.Fatalf("%+v", u)
	}
}

func TestClientDepositAlreadyL2(t *testing.T) {
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		if method != "getPoint" {
			return nil, fmt.Errorf("unexpected %s", method)
		}
		return testPointResult(DominionL2), nil
	})
	if _, err := c.Deposit(context.Background(), 256); err == nil {
		t.Fatal("expected error")
	}
}

func TestClientDepositGalaxy(t *testing.T) {
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		if method != "getPoint" {
			return nil, fmt.Errorf("unexpected %s", method)
		}
		return testPointResult(DominionL1), nil
	})
	if _, err := c.Deposit(context.Background(), 69); err == nil {
		t.Fatal("expected error")
	}
}

func TestClientTransferPointL2(t *testing.T) {
	target := common.HexToAddress("0xd1428F18A8255C0984291CEBFc83E6F982F7De9f")
	from := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	var sawTx string
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		switch method {
		case "getPoint":
			res := Result{Dominion: DominionL2}
			res.Ownership.Owner.Address = from.Hex()
			res.Ownership.Owner.Nonce = 2
			return res, nil
		case "prepareForSigning":
			var p struct {
				Tx string `json:"tx"`
			}
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, err
			}
			sawTx = p.Tx
			return "0xtr", nil
		default:
			return nil, fmt.Errorf("unexpected %s", method)
		}
	})

	u, err := c.TransferPoint(context.Background(), 69, target, false, from)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL2 || u.Hash != "0xtr" || sawTx != L2TxTransferPoint {
		t.Fatalf("layer=%s hash=%s tx=%s", u.Layer, u.Hash, sawTx)
	}
}

func TestClientEscapeL1(t *testing.T) {
	from := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		if method != "getPoint" {
			return nil, fmt.Errorf("unexpected %s", method)
		}
		return Result{Dominion: DominionL1}, nil
	})

	u, err := c.Escape(context.Background(), 69, 256, from)
	if err != nil {
		t.Fatal(err)
	}
	want, err := PackEscape(69, 256)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL1 || !bytes.Equal(u.Data, want) {
		t.Fatalf("%+v", u)
	}
}

func TestClientEscapeSpawnDominion(t *testing.T) {
	from := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	var saw struct {
		Tx   string   `json:"tx"`
		From L2From   `json:"from"`
		Data ShipData `json:"data"`
	}
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		switch method {
		case "getPoint":
			res := Result{Dominion: DominionSpawn}
			res.Ownership.Owner.Address = from.Hex()
			res.Ownership.Owner.Nonce = 8
			return res, nil
		case "prepareForSigning":
			if err := json.Unmarshal(params, &saw); err != nil {
				return nil, err
			}
			return "0xesc", nil
		default:
			return nil, fmt.Errorf("unexpected %s", method)
		}
	})

	u, err := c.Escape(context.Background(), 256, 0, from)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL2 || u.Hash != "0xesc" || saw.Tx != L2TxEscape {
		t.Fatalf("%+v saw=%+v", u, saw)
	}
	if saw.From.Ship != "~marzod" || saw.Data.Ship != "~zod" {
		t.Fatalf("from %+v data %+v", saw.From, saw.Data)
	}
}

func TestClientAdoptL1(t *testing.T) {
	from := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		if method != "getPoint" {
			return nil, fmt.Errorf("unexpected %s", method)
		}
		return Result{Dominion: DominionL1}, nil
	})

	u, err := c.Adopt(context.Background(), 256, 69, from)
	if err != nil {
		t.Fatal(err)
	}
	want, err := PackAdopt(69)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL1 || !bytes.Equal(u.Data, want) {
		t.Fatalf("%+v", u)
	}
}

func TestClientAdoptL2(t *testing.T) {
	from := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	var saw struct {
		Tx   string   `json:"tx"`
		From L2From   `json:"from"`
		Data ShipData `json:"data"`
	}
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		switch method {
		case "getPoint":
			res := Result{Dominion: DominionL2}
			res.Ownership.Owner.Address = from.Hex()
			res.Ownership.Owner.Nonce = 1
			return res, nil
		case "prepareForSigning":
			if err := json.Unmarshal(params, &saw); err != nil {
				return nil, err
			}
			return "0xad", nil
		default:
			return nil, fmt.Errorf("unexpected %s", method)
		}
	})

	u, err := c.Adopt(context.Background(), 256, 69, from)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL2 || saw.Tx != L2TxAdopt || saw.From.Ship != "~marzod" || saw.Data.Ship != "~pet" {
		t.Fatalf("%+v saw=%+v", u, saw)
	}
}

func TestClientSubmitL2(t *testing.T) {
	from := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	sig := "0x" + strings.Repeat("ab", 65)
	mgr := common.HexToAddress("0xF7306c5db0C1880FB2ed9c3972ad3e1A94999196")
	ctx := context.Background()

	l2Owner := func() Result {
		res := Result{Dominion: DominionL2}
		res.Ownership.Owner.Address = from.Hex()
		res.Ownership.Owner.Nonce = 2
		return res
	}

	t.Run("setManagementProxy", func(t *testing.T) {
		c := testClient(t, func(method string, params json.RawMessage) (any, error) {
			switch method {
			case "getPoint":
				return l2Owner(), nil
			case "prepareForSigning":
				return "0xcafe", nil
			case L2TxSetManagementProxy:
				var p struct {
					Sig     string `json:"sig"`
					Address string `json:"address"`
				}
				if err := json.Unmarshal(params, &p); err != nil {
					return nil, err
				}
				if p.Sig != sig || !strings.EqualFold(p.Address, from.Hex()) {
					return nil, fmt.Errorf("params %+v", p)
				}
				return "0xbeef", nil
			default:
				return nil, fmt.Errorf("unexpected %s", method)
			}
		})
		u, err := c.SetManagementProxy(ctx, 69, mgr, from)
		if err != nil {
			t.Fatal(err)
		}
		got, err := c.SubmitL2(ctx, u, sig, from)
		if err != nil {
			t.Fatal(err)
		}
		if got != "0xbeef" {
			t.Fatalf("hash %q", got)
		}
		if c.Roller() == nil {
			t.Fatal("nil roller")
		}
	})

	t.Run("spawn", func(t *testing.T) {
		c := testClient(t, func(method string, params json.RawMessage) (any, error) {
			switch method {
			case "getPoint":
				res := Result{Dominion: DominionSpawn}
				res.Ownership.Owner.Address = from.Hex()
				res.Ownership.Owner.Nonce = 4
				return res, nil
			case "prepareForSigning":
				return "0xsp", nil
			case L2TxSpawn:
				return "0xbeef", nil
			default:
				return nil, fmt.Errorf("unexpected %s", method)
			}
		})
		u, err := c.Spawn(ctx, 1566792653, mgr, from)
		if err != nil {
			t.Fatal(err)
		}
		got, err := c.SubmitL2(ctx, u, sig, from)
		if err != nil {
			t.Fatal(err)
		}
		if got != "0xbeef" {
			t.Fatalf("hash %q", got)
		}
	})

	t.Run("escape", func(t *testing.T) {
		c := testClient(t, func(method string, params json.RawMessage) (any, error) {
			switch method {
			case "getPoint":
				res := Result{Dominion: DominionSpawn}
				res.Ownership.Owner.Address = from.Hex()
				res.Ownership.Owner.Nonce = 8
				return res, nil
			case "prepareForSigning":
				return "0xesc", nil
			case L2TxEscape:
				return "0xbeef", nil
			default:
				return nil, fmt.Errorf("unexpected %s", method)
			}
		})
		u, err := c.Escape(ctx, 256, 0, from)
		if err != nil {
			t.Fatal(err)
		}
		got, err := c.SubmitL2(ctx, u, sig, from)
		if err != nil {
			t.Fatal(err)
		}
		if got != "0xbeef" {
			t.Fatalf("hash %q", got)
		}
	})
}

func TestClientAddClaim(t *testing.T) {
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		return nil, fmt.Errorf("unexpected %s", method)
	})

	u, err := c.AddClaim(context.Background(), 69, "prot1", "claim", []byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	want, err := PackAddClaim(69, "prot1", "claim", []byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL1 || u.To != ClaimsAddr() || !bytes.Equal(u.Data, want) {
		t.Fatalf("%+v", u)
	}
}

func TestClientRemoveClaim(t *testing.T) {
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		return nil, fmt.Errorf("unexpected %s", method)
	})
	u, err := c.RemoveClaim(context.Background(), 69, "prot1", "claim")
	if err != nil {
		t.Fatal(err)
	}
	want, err := PackRemoveClaim(69, "prot1", "claim")
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL1 || u.To != ClaimsAddr() || !bytes.Equal(u.Data, want) {
		t.Fatalf("%+v", u)
	}
}

func TestClientClearClaims(t *testing.T) {
	c := testClient(t, func(method string, params json.RawMessage) (any, error) {
		return nil, fmt.Errorf("unexpected %s", method)
	})
	u, err := c.ClearClaims(context.Background(), 69)
	if err != nil {
		t.Fatal(err)
	}
	want, err := PackClearClaims(69)
	if err != nil {
		t.Fatal(err)
	}
	if u.Layer != LayerL1 || u.To != ClaimsAddr() || !bytes.Equal(u.Data, want) {
		t.Fatalf("%+v", u)
	}
}

func TestClientAddClaimLiveAddress(t *testing.T) {
	ec, err := NewEclipticContract()
	if err != nil {
		t.Fatal(err)
	}
	live := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	out, err := ec.ABI.Methods["claims"].Outputs.Pack(live)
	if err != nil {
		t.Fatal(err)
	}
	claimsSel := hexutil.Encode(mustPack(t, ec, "claims")[:4])

	eth := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		switch method {
		case "eth_chainId", "net_version":
			return "0x1", nil
		case "eth_call":
			data := ethCallData(params)
			if strings.HasPrefix(data, claimsSel) {
				return hexutil.Encode(out), nil
			}
			return nil, fmt.Errorf("unexpected call %s", data)
		default:
			return nil, fmt.Errorf("method not found: %s", method)
		}
	})
	defer eth.Close()
	roller := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		return nil, fmt.Errorf("unexpected %s", method)
	})
	defer roller.Close()

	c, err := Dial(eth.URL, roller.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	u, err := c.AddClaim(context.Background(), 69, "prot1", "claim", nil)
	if err != nil {
		t.Fatal(err)
	}
	if u.To != live {
		t.Fatalf("to %s want %s", u.To.Hex(), live.Hex())
	}
}

func TestClientGetClaims(t *testing.T) {
	cl, err := NewClaimsContract()
	if err != nil {
		t.Fatal(err)
	}
	filled, err := cl.ABI.Methods["claims"].Outputs.Pack("prot1", "claim", []byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	empty, err := cl.ABI.Methods["claims"].Outputs.Pack("", "", []byte{})
	if err != nil {
		t.Fatal(err)
	}
	claimsSel := hexutil.Encode(mustPack(t, cl, "claims", uint32(69), big.NewInt(0))[:4])

	eth := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		switch method {
		case "eth_chainId", "net_version":
			return "0x1", nil
		case "eth_call":
			data := ethCallData(params)
			if !strings.HasPrefix(data, claimsSel) {
				return nil, fmt.Errorf("unexpected call %s", data)
			}
			idx := claimIndexFromCall(t, cl, data)
			if idx == 1 {
				return hexutil.Encode(filled), nil
			}
			return hexutil.Encode(empty), nil
		default:
			return nil, fmt.Errorf("method not found: %s", method)
		}
	})
	defer eth.Close()
	roller := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		return nil, fmt.Errorf("unexpected %s", method)
	})
	defer roller.Close()

	c, err := Dial(eth.URL, roller.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	got, err := c.GetClaims(context.Background(), 69)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Protocol != "prot1" || got[0].Claim != "claim" {
		t.Fatalf("%+v", got)
	}

	one, err := c.GetClaim(context.Background(), 69, 1)
	if err != nil {
		t.Fatal(err)
	}
	if one.Protocol != "prot1" {
		t.Fatalf("%+v", one)
	}
}

func TestClientGetClaimsNoEth(t *testing.T) {
	if _, err := (*Client)(nil).GetClaims(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestClientSubmitL2NoRoller(t *testing.T) {
	_, err := (*Client)(nil).SubmitL2(context.Background(), &Unsigned{Layer: LayerL2}, "0x00", common.Address{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func getPointL1Fallback(t *testing.T, az *Contract, pointsOut, rightsOut []byte) *Point {
	t.Helper()
	pointsSel := hexutil.Encode(mustPack(t, az, "points", uint32(0))[:4])
	rightsSel := hexutil.Encode(mustPack(t, az, "rights", uint32(0))[:4])
	eth := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		switch method {
		case "eth_chainId", "net_version":
			return "0x1", nil
		case "eth_call":
			data := ethCallData(params)
			switch {
			case strings.HasPrefix(data, pointsSel):
				return hexutil.Encode(pointsOut), nil
			case strings.HasPrefix(data, rightsSel):
				return hexutil.Encode(rightsOut), nil
			default:
				return nil, fmt.Errorf("unexpected call %s", data)
			}
		default:
			return nil, fmt.Errorf("method not found: %s", method)
		}
	})
	t.Cleanup(eth.Close)
	roller := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		return nil, fmt.Errorf("roller down")
	})
	t.Cleanup(roller.Close)
	c, err := Dial(eth.URL, roller.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	p, err := c.GetPoint(context.Background(), 69)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func testPointResult(dominion string) Result {
	res := Result{Dominion: dominion}
	res.Network.Keys.Life = "1"
	res.Network.Keys.Suite = "1"
	res.Network.Keys.Auth = "0x" + strings.Repeat("11", 32)
	res.Network.Keys.Crypt = "0x" + strings.Repeat("22", 32)
	return res
}

func testClient(t *testing.T, handle func(method string, params json.RawMessage) (any, error)) *Client {
	t.Helper()
	roller := mockJSONRPC(t, handle)
	t.Cleanup(roller.Close)
	eth := mockEthRPC(t, nil)
	t.Cleanup(eth.Close)
	c, err := Dial(eth.URL, roller.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	return c
}

func mustPack(t *testing.T, c *Contract, name string, args ...any) []byte {
	t.Helper()
	data, err := c.Pack(name, args...)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
