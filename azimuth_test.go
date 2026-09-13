package azimuth

import (
	"bytes"
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestPackUnpackGetOwnedPoints(t *testing.T) {
	c, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	addr := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	in, err := c.Pack("getOwnedPoints", addr)
	if err != nil {
		t.Fatal(err)
	}
	sel := crypto.Keccak256([]byte("getOwnedPoints(address)"))[:4]
	if !bytes.Equal(in[:4], sel) {
		t.Fatalf("selector %x want %x", in[:4], sel)
	}

	out, err := c.ABI.Methods["getOwnedPoints"].Outputs.Pack([]uint32{0, 256, 69})
	if err != nil {
		t.Fatal(err)
	}
	unpacked, err := c.ABI.Unpack("getOwnedPoints", out)
	if err != nil {
		t.Fatal(err)
	}
	got, err := unpackUint32s(unpacked[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0] != 0 || got[1] != 256 || got[2] != 69 {
		t.Fatalf("got %v", got)
	}
}

func TestPackUnpackCanManage(t *testing.T) {
	c, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	addr := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	in, err := c.Pack("canManage", uint32(69), addr)
	if err != nil {
		t.Fatal(err)
	}
	sel := crypto.Keccak256([]byte("canManage(uint32,address)"))[:4]
	if !bytes.Equal(in[:4], sel) {
		t.Fatalf("selector %x", in[:4])
	}
	out, err := c.ABI.Methods["canManage"].Outputs.Pack(true)
	if err != nil {
		t.Fatal(err)
	}
	unpacked, err := c.ABI.Unpack("canManage", out)
	if err != nil {
		t.Fatal(err)
	}
	ok, isBool := unpacked[0].(bool)
	if !isBool || !ok {
		t.Fatalf("unpacked %v", unpacked[0])
	}
}

func TestPackUnpackGetOwner(t *testing.T) {
	c, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	want := common.HexToAddress(EclipticAddress)
	out, err := c.ABI.Methods["getOwner"].Outputs.Pack(want)
	if err != nil {
		t.Fatal(err)
	}
	unpacked, err := c.ABI.Unpack("getOwner", out)
	if err != nil {
		t.Fatal(err)
	}
	got, err := unpackAddress(unpacked[0])
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestGetOwnedPointsMockRPC(t *testing.T) {
	c, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.ABI.Methods["getOwnedPoints"].Outputs.Pack([]uint32{69, 256})
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
	got, err := GetOwnedPoints(context.Background(), client, common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != 69 || got[1] != 256 {
		t.Fatalf("got %v", got)
	}
}

func TestCanManageMockRPC(t *testing.T) {
	c, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.ABI.Methods["canManage"].Outputs.Pack(true)
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
	ok, err := CanManage(context.Background(), client, 69, common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F"))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("want true")
	}
}

func TestGetKeysMockRPC(t *testing.T) {
	c, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	var crypt, auth [32]byte
	copy(crypt[:], bytes.Repeat([]byte{0xaa}, 32))
	copy(auth[:], bytes.Repeat([]byte{0xbb}, 32))
	out, err := c.ABI.Methods["getKeys"].Outputs.Pack(crypt, auth, uint32(1), uint32(3))
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
	keys, err := GetKeys(context.Background(), client, 69)
	if err != nil {
		t.Fatal(err)
	}
	if keys.Suite != 1 || keys.Revision != 3 {
		t.Fatalf("%+v", keys)
	}
	if !bytes.Equal(keys.CryptKey, crypt[:]) || !bytes.Equal(keys.EdKey, auth[:]) {
		t.Fatalf("keys %+v", keys)
	}
}

func TestOwnerMockRPC(t *testing.T) {
	c, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	want := EclipticAddr()
	out, err := c.ABI.Methods["owner"].Outputs.Pack(want)
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
	got, err := Owner(context.Background(), client)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %s want %s", got.Hex(), want.Hex())
	}
}

func TestUnpackUint32sEmpty(t *testing.T) {
	got, err := unpackUint32s([]uint32{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestGetPointDataMockRPC(t *testing.T) {
	c, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	var crypt, auth [32]byte
	copy(crypt[:], bytes.Repeat([]byte{0xaa}, 32))
	copy(auth[:], bytes.Repeat([]byte{0xbb}, 32))
	out, err := c.ABI.Methods["points"].Outputs.Pack(
		crypt, auth, true, true, true, uint32(256), uint32(512), uint32(1), uint32(3), uint32(7),
	)
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
	got, err := GetPointData(context.Background(), client, 69)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.CryptKey, crypt[:]) || !bytes.Equal(got.EdKey, auth[:]) {
		t.Fatalf("keys %+v", got)
	}
	if !got.HasSponsor || !got.Active || !got.EscapeRequested {
		t.Fatalf("flags %+v", got)
	}
	if got.Sponsor != 256 || got.EscapeRequestedTo != 512 || got.Suite != 1 || got.Revision != 3 || got.Rift != 7 {
		t.Fatalf("%+v", got)
	}
}

func TestGetRightsMockRPC(t *testing.T) {
	c, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	owner := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	mgmt := common.HexToAddress("0xF7306c5db0C1880FB2ed9c3972ad3e1A94999196")
	spawn := common.HexToAddress("0xd1428F18A8255C0984291CEBFc83E6F982F7De9f")
	vote := common.HexToAddress("0x1111111111111111111111111111111111111112")
	xfer := common.HexToAddress("0x2222222222222222222222222222222222222222")
	out, err := c.ABI.Methods["rights"].Outputs.Pack(owner, mgmt, spawn, vote, xfer)
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
	got, err := GetRights(context.Background(), client, 69)
	if err != nil {
		t.Fatal(err)
	}
	if got.Owner != owner || got.ManagementProxy != mgmt || got.SpawnProxy != spawn || got.VotingProxy != vote || got.TransferProxy != xfer {
		t.Fatalf("%+v", got)
	}
}

func TestGetListsMockRPC(t *testing.T) {
	c, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	addr := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	want := []uint32{69, 256}
	tests := []struct {
		method string
		get    func(context.Context, *ethclient.Client) ([]uint32, error)
	}{
		{"getSpawned", func(ctx context.Context, client *ethclient.Client) ([]uint32, error) {
			return GetSpawned(ctx, client, 0)
		}},
		{"getSponsoring", func(ctx context.Context, client *ethclient.Client) ([]uint32, error) {
			return GetSponsoring(ctx, client, 0)
		}},
		{"getEscapeRequests", func(ctx context.Context, client *ethclient.Client) ([]uint32, error) {
			return GetEscapeRequests(ctx, client, 0)
		}},
		{"getSpawningFor", func(ctx context.Context, client *ethclient.Client) ([]uint32, error) {
			return GetSpawningFor(ctx, client, addr)
		}},
		{"getTransferringFor", func(ctx context.Context, client *ethclient.Client) ([]uint32, error) {
			return GetTransferringFor(ctx, client, addr)
		}},
		{"getVotingFor", func(ctx context.Context, client *ethclient.Client) ([]uint32, error) {
			return GetVotingFor(ctx, client, addr)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			out, err := c.ABI.Methods[tc.method].Outputs.Pack(want)
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
			got, err := tc.get(context.Background(), client)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 2 || got[0] != 69 || got[1] != 256 {
				t.Fatalf("got %v", got)
			}
		})
	}
}
