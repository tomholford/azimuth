package azimuth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

//
// Constants
//

// MaxClaims is Claims.sol maxClaims — 16 slots per point.
const MaxClaims = 16

//
// Types
//

// Claim is one Claims.sol slot: protocol (context), claim, and dossier (proof).
type Claim struct {
	Protocol string
	Claim    string
	Dossier  []byte
}

//
// Public
//

func PackAddClaim(point uint32, protocol, claim string, dossier []byte) ([]byte, error) {
	if err := requireClaimFields(protocol, claim); err != nil {
		return nil, fmt.Errorf("addClaim: %w", err)
	}
	if dossier == nil {
		dossier = []byte{}
	}
	return packClaims("addClaim", point, protocol, claim, dossier)
}

func PackRemoveClaim(point uint32, protocol, claim string) ([]byte, error) {
	if err := requireClaimFields(protocol, claim); err != nil {
		return nil, fmt.Errorf("removeClaim: %w", err)
	}
	return packClaims("removeClaim", point, protocol, claim)
}

func PackClearClaims(point uint32) ([]byte, error) {
	return packClaims("clearClaims", point)
}

func GetClaim(ctx context.Context, client *ethclient.Client, point uint32, index uint8) (*Claim, error) {
	if index >= MaxClaims {
		return nil, fmt.Errorf("getClaim: index %d >= %d", index, MaxClaims)
	}
	out, err := claimsCall(ctx, client, "claims", point, big.NewInt(int64(index)))
	if err != nil {
		return nil, err
	}
	return unpackClaim(out)
}

// FindClaim is Claims.findClaim: 0 if missing, slot index + 1 otherwise.
func FindClaim(ctx context.Context, client *ethclient.Client, point uint32, protocol, claim string) (uint8, error) {
	if err := requireClaimFields(protocol, claim); err != nil {
		return 0, fmt.Errorf("findClaim: %w", err)
	}
	out, err := claimsCall(ctx, client, "findClaim", point, protocol, claim)
	if err != nil {
		return 0, err
	}
	if len(out) != 1 {
		return 0, fmt.Errorf("findClaim: unexpected result count %d", len(out))
	}
	n, err := unpackUint32(out[0])
	if err != nil {
		return 0, fmt.Errorf("findClaim: %w", err)
	}
	if n > uint32(^uint8(0)) {
		return 0, fmt.Errorf("findClaim: index %d overflows uint8", n)
	}
	return uint8(n), nil
}

// GetClaims reads all 16 slots and drops empty ones (both protocol and claim blank).
func GetClaims(ctx context.Context, client *ethclient.Client, point uint32) ([]Claim, error) {
	out := make([]Claim, 0)
	for i := uint8(0); i < MaxClaims; i++ {
		c, err := GetClaim(ctx, client, point, i)
		if err != nil {
			return nil, err
		}
		if c.Protocol == "" && c.Claim == "" {
			continue
		}
		out = append(out, *c)
	}
	return out, nil
}

//
// Private
//

func packClaims(name string, args ...any) ([]byte, error) {
	c, err := NewClaimsContract()
	if err != nil {
		return nil, err
	}
	return c.Pack(name, args...)
}

func claimsCall(ctx context.Context, client *ethclient.Client, name string, args ...any) ([]any, error) {
	c, err := NewClaimsContract()
	if err != nil {
		return nil, err
	}
	return c.Call(ctx, client, name, args...)
}

func requireClaimFields(protocol, claim string) error {
	if protocol == "" || claim == "" {
		return fmt.Errorf("empty protocol or claim")
	}
	return nil
}

func unpackClaim(out []any) (*Claim, error) {
	if len(out) != 3 {
		return nil, fmt.Errorf("claim: unexpected result count %d", len(out))
	}
	protocol, err := unpackString(out[0])
	if err != nil {
		return nil, fmt.Errorf("claim protocol: %w", err)
	}
	claim, err := unpackString(out[1])
	if err != nil {
		return nil, fmt.Errorf("claim: %w", err)
	}
	dossier, err := unpackBytes(out[2])
	if err != nil {
		return nil, fmt.Errorf("claim dossier: %w", err)
	}
	return &Claim{Protocol: protocol, Claim: claim, Dossier: dossier}, nil
}
