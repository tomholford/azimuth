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

// PointData is Azimuth.points — keys, sponsor, escape, continuity.
type PointData struct {
	CryptKey          []byte
	EdKey             []byte
	HasSponsor        bool
	Active            bool
	EscapeRequested   bool
	Sponsor           uint32
	EscapeRequestedTo uint32
	Suite             uint32
	Revision          uint32
	Rift              uint32
}

// Deed is Azimuth.rights — owner and proxies.
type Deed struct {
	Owner           common.Address
	ManagementProxy common.Address
	SpawnProxy      common.Address
	VotingProxy     common.Address
	TransferProxy   common.Address
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
	return azimuthUint32s(ctx, client, "getOwnedPoints", whose)
}

func GetManagerFor(ctx context.Context, client *ethclient.Client, whose common.Address) ([]uint32, error) {
	return azimuthUint32s(ctx, client, "getManagerFor", whose)
}

func GetSpawningFor(ctx context.Context, client *ethclient.Client, whose common.Address) ([]uint32, error) {
	return azimuthUint32s(ctx, client, "getSpawningFor", whose)
}

func GetTransferringFor(ctx context.Context, client *ethclient.Client, whose common.Address) ([]uint32, error) {
	return azimuthUint32s(ctx, client, "getTransferringFor", whose)
}

func GetVotingFor(ctx context.Context, client *ethclient.Client, whose common.Address) ([]uint32, error) {
	return azimuthUint32s(ctx, client, "getVotingFor", whose)
}

func GetSpawned(ctx context.Context, client *ethclient.Client, point uint32) ([]uint32, error) {
	return azimuthUint32s(ctx, client, "getSpawned", point)
}

func GetSponsoring(ctx context.Context, client *ethclient.Client, point uint32) ([]uint32, error) {
	return azimuthUint32s(ctx, client, "getSponsoring", point)
}

func GetEscapeRequests(ctx context.Context, client *ethclient.Client, point uint32) ([]uint32, error) {
	return azimuthUint32s(ctx, client, "getEscapeRequests", point)
}

func GetPointData(ctx context.Context, client *ethclient.Client, point uint32) (*PointData, error) {
	out, err := azimuthCall(ctx, client, "points", point)
	if err != nil {
		return nil, err
	}
	return unpackPointData(out)
}

func GetRights(ctx context.Context, client *ethclient.Client, point uint32) (*Deed, error) {
	out, err := azimuthCall(ctx, client, "rights", point)
	if err != nil {
		return nil, err
	}
	return unpackDeed(out)
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
	ok, err := unpackBool(out[0])
	if err != nil {
		return false, fmt.Errorf("canManage: %w", err)
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

func azimuthUint32s(ctx context.Context, client *ethclient.Client, name string, args ...any) ([]uint32, error) {
	out, err := azimuthCall(ctx, client, name, args...)
	if err != nil {
		return nil, err
	}
	if len(out) != 1 {
		return nil, fmt.Errorf("%s: unexpected result count %d", name, len(out))
	}
	return unpackUint32s(out[0])
}

func unpackPointData(out []any) (*PointData, error) {
	if len(out) != 10 {
		return nil, fmt.Errorf("points: unexpected result count %d", len(out))
	}
	crypt, err := unpackBytes32(out[0])
	if err != nil {
		return nil, fmt.Errorf("points crypt: %w", err)
	}
	auth, err := unpackBytes32(out[1])
	if err != nil {
		return nil, fmt.Errorf("points auth: %w", err)
	}
	hasSponsor, err := unpackBool(out[2])
	if err != nil {
		return nil, fmt.Errorf("points hasSponsor: %w", err)
	}
	active, err := unpackBool(out[3])
	if err != nil {
		return nil, fmt.Errorf("points active: %w", err)
	}
	escaping, err := unpackBool(out[4])
	if err != nil {
		return nil, fmt.Errorf("points escapeRequested: %w", err)
	}
	sponsor, err := unpackUint32(out[5])
	if err != nil {
		return nil, fmt.Errorf("points sponsor: %w", err)
	}
	escapeTo, err := unpackUint32(out[6])
	if err != nil {
		return nil, fmt.Errorf("points escapeRequestedTo: %w", err)
	}
	suite, err := unpackUint32(out[7])
	if err != nil {
		return nil, fmt.Errorf("points suite: %w", err)
	}
	rev, err := unpackUint32(out[8])
	if err != nil {
		return nil, fmt.Errorf("points revision: %w", err)
	}
	rift, err := unpackUint32(out[9])
	if err != nil {
		return nil, fmt.Errorf("points continuity: %w", err)
	}
	return &PointData{
		CryptKey:          crypt,
		EdKey:             auth,
		HasSponsor:        hasSponsor,
		Active:            active,
		EscapeRequested:   escaping,
		Sponsor:           sponsor,
		EscapeRequestedTo: escapeTo,
		Suite:             suite,
		Revision:          rev,
		Rift:              rift,
	}, nil
}

func unpackDeed(out []any) (*Deed, error) {
	if len(out) != 5 {
		return nil, fmt.Errorf("rights: unexpected result count %d", len(out))
	}
	owner, err := unpackAddress(out[0])
	if err != nil {
		return nil, fmt.Errorf("rights owner: %w", err)
	}
	mgmt, err := unpackAddress(out[1])
	if err != nil {
		return nil, fmt.Errorf("rights managementProxy: %w", err)
	}
	spawn, err := unpackAddress(out[2])
	if err != nil {
		return nil, fmt.Errorf("rights spawnProxy: %w", err)
	}
	vote, err := unpackAddress(out[3])
	if err != nil {
		return nil, fmt.Errorf("rights votingProxy: %w", err)
	}
	xfer, err := unpackAddress(out[4])
	if err != nil {
		return nil, fmt.Errorf("rights transferProxy: %w", err)
	}
	return &Deed{
		Owner:           owner,
		ManagementProxy: mgmt,
		SpawnProxy:      spawn,
		VotingProxy:     vote,
		TransferProxy:   xfer,
	}, nil
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

func unpackString(v any) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("unexpected string type %T", v)
	}
	return s, nil
}

func unpackBytes(v any) ([]byte, error) {
	switch b := v.(type) {
	case []byte:
		return b, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("unexpected bytes type %T", v)
	}
}

func unpackBool(v any) (bool, error) {
	ok, isBool := v.(bool)
	if !isBool {
		return false, fmt.Errorf("unexpected type %T", v)
	}
	return ok, nil
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
