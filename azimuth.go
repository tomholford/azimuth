package azimuth

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

//
// Types
//

type Keys struct {
	CryptKey []byte
	EdKey    []byte
	Suite    uint32
	Revision uint32
}

//
// Public
//

func GetKeys(ctx context.Context, client *ethclient.Client, point uint32) (*Keys, error) {
	out, err := azimuthCall(ctx, client, "getKeys", point)
	if err != nil {
		return nil, err
	}
	if len(out) != 4 {
		return nil, fmt.Errorf("getKeys: unexpected result count %d", len(out))
	}
	crypt, err := unpackBytes32(out[0])
	if err != nil {
		return nil, fmt.Errorf("getKeys crypt: %w", err)
	}
	auth, err := unpackBytes32(out[1])
	if err != nil {
		return nil, fmt.Errorf("getKeys auth: %w", err)
	}
	suite, err := unpackUint32(out[2])
	if err != nil {
		return nil, fmt.Errorf("getKeys suite: %w", err)
	}
	rev, err := unpackUint32(out[3])
	if err != nil {
		return nil, fmt.Errorf("getKeys revision: %w", err)
	}
	return &Keys{CryptKey: crypt, EdKey: auth, Suite: suite, Revision: rev}, nil
}

func GetOwnedPoints(ctx context.Context, client *ethclient.Client, whose common.Address) ([]uint32, error) {
	out, err := azimuthCall(ctx, client, "getOwnedPoints", whose)
	if err != nil {
		return nil, err
	}
	if len(out) != 1 {
		return nil, fmt.Errorf("getOwnedPoints: unexpected result count %d", len(out))
	}
	return unpackUint32s(out[0])
}

func GetManagerFor(ctx context.Context, client *ethclient.Client, whose common.Address) ([]uint32, error) {
	out, err := azimuthCall(ctx, client, "getManagerFor", whose)
	if err != nil {
		return nil, err
	}
	if len(out) != 1 {
		return nil, fmt.Errorf("getManagerFor: unexpected result count %d", len(out))
	}
	return unpackUint32s(out[0])
}

func GetOwner(ctx context.Context, client *ethclient.Client, point uint32) (common.Address, error) {
	out, err := azimuthCall(ctx, client, "getOwner", point)
	if err != nil {
		return common.Address{}, err
	}
	if len(out) != 1 {
		return common.Address{}, fmt.Errorf("getOwner: unexpected result count %d", len(out))
	}
	return unpackAddress(out[0])
}

func CanManage(ctx context.Context, client *ethclient.Client, point uint32, whose common.Address) (bool, error) {
	out, err := azimuthCall(ctx, client, "canManage", point, whose)
	if err != nil {
		return false, err
	}
	if len(out) != 1 {
		return false, fmt.Errorf("canManage: unexpected result count %d", len(out))
	}
	ok, isBool := out[0].(bool)
	if !isBool {
		return false, fmt.Errorf("canManage: unexpected type %T", out[0])
	}
	return ok, nil
}

// Owner is the Azimuth contract owner — the current Ecliptic.
func Owner(ctx context.Context, client *ethclient.Client) (common.Address, error) {
	out, err := azimuthCall(ctx, client, "owner")
	if err != nil {
		return common.Address{}, err
	}
	if len(out) != 1 {
		return common.Address{}, fmt.Errorf("owner: unexpected result count %d", len(out))
	}
	return unpackAddress(out[0])
}

//
// Private
//

func azimuthCall(ctx context.Context, client *ethclient.Client, name string, args ...any) ([]any, error) {
	c, err := NewAzimuthContract()
	if err != nil {
		return nil, err
	}
	return c.Call(ctx, client, name, args...)
}

func unpackAddress(v any) (common.Address, error) {
	addr, ok := v.(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("unexpected address type %T", v)
	}
	return addr, nil
}

func unpackBytes32(v any) ([]byte, error) {
	switch b := v.(type) {
	case [32]byte:
		out := make([]byte, 32)
		copy(out, b[:])
		return out, nil
	case []byte:
		if len(b) != 32 {
			return nil, fmt.Errorf("want 32 bytes, got %d", len(b))
		}
		return b, nil
	default:
		return nil, fmt.Errorf("unexpected bytes32 type %T", v)
	}
}

func unpackUint32s(v any) ([]uint32, error) {
	switch a := v.(type) {
	case []uint32:
		return a, nil
	case []any:
		out := make([]uint32, len(a))
		for i, x := range a {
			n, err := unpackUint32(x)
			if err != nil {
				return nil, fmt.Errorf("element %d: %w", i, err)
			}
			out[i] = n
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unexpected uint32[] type %T", v)
	}
}

func unpackUint32(v any) (uint32, error) {
	switch n := v.(type) {
	case uint32:
		return n, nil
	case uint16:
		return uint32(n), nil
	case uint8:
		return uint32(n), nil
	case uint64:
		if n > uint64(^uint32(0)) {
			return 0, fmt.Errorf("uint64 %d overflows uint32", n)
		}
		return uint32(n), nil
	default:
		return 0, fmt.Errorf("unexpected type %T", v)
	}
}
