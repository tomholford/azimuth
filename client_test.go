package azimuth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	keysOut, err := az.ABI.Methods["getKeys"].Outputs.Pack(crypt, auth, uint32(1), uint32(3))
	if err != nil {
		t.Fatal(err)
	}
	owner := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	ownerOut, err := az.ABI.Methods["getOwner"].Outputs.Pack(owner)
	if err != nil {
		t.Fatal(err)
	}
	keysSel := hexutil.Encode(mustPack(t, az, "getKeys", uint32(0))[:4])
	ownerSel := hexutil.Encode(mustPack(t, az, "getOwner", uint32(0))[:4])

	eth := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		switch method {
		case "eth_chainId", "net_version":
			return "0x1", nil
		case "eth_call":
			data := ethCallData(params)
			switch {
			case strings.HasPrefix(data, keysSel):
				return hexutil.Encode(keysOut), nil
			case strings.HasPrefix(data, ownerSel):
				return hexutil.Encode(ownerOut), nil
			default:
				return nil, fmt.Errorf("unexpected call %s", data)
			}
		default:
			return nil, fmt.Errorf("method not found: %s", method)
		}
	})
	defer eth.Close()
	roller := mockJSONRPC(t, func(method string, params json.RawMessage) (any, error) {
		return nil, fmt.Errorf("roller down")
	})
	defer roller.Close()

	c, err := Dial(eth.URL, roller.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	p, err := c.GetPoint(context.Background(), 69)
	if err != nil {
		t.Fatal(err)
	}
	if p.Dominion != DominionL1 || p.Owner.Address != owner || p.Revision != 3 {
		t.Fatalf("%+v", p)
	}
	if !bytes.Equal(p.CryptKey, crypt[:]) {
		t.Fatalf("crypt %x", p.CryptKey)
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

func TestClientSubmitL2NoRoller(t *testing.T) {
	_, err := (*Client)(nil).SubmitL2(context.Background(), &Unsigned{Layer: LayerL2}, "0x00", common.Address{})
	if err == nil {
		t.Fatal("expected error")
	}
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
