package azimuth

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

//
// Constants
//

// CryptoSuiteVersion is Bridge CRYPTO_SUITE_VERSION (Urbit OS suite 1).
const CryptoSuiteVersion uint32 = 1

//
// Public
//

func PackConfigureKeys(point uint32, encryptionKey, authenticationKey []byte, suite uint32, discontinuous bool) ([]byte, error) {
	enc, err := asBytes32(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("encryptionKey: %w", err)
	}
	auth, err := asBytes32(authenticationKey)
	if err != nil {
		return nil, fmt.Errorf("authenticationKey: %w", err)
	}
	return packEcliptic("configureKeys", point, enc, auth, suite, discontinuous)
}

func PackSetManagementProxy(point uint32, manager common.Address) ([]byte, error) {
	return packEcliptic("setManagementProxy", point, manager)
}

func PackSetSpawnProxy(point uint32, spawnProxy common.Address) ([]byte, error) {
	if point > 0xffff {
		return nil, fmt.Errorf("setSpawnProxy: point %d is not a star or galaxy", point)
	}
	return packEcliptic("setSpawnProxy", uint16(point), spawnProxy)
}

func PackSetTransferProxy(point uint32, transferProxy common.Address) ([]byte, error) {
	return packEcliptic("setTransferProxy", point, transferProxy)
}

func PackSetVotingProxy(point uint32, votingProxy common.Address) ([]byte, error) {
	if point > 0xff {
		return nil, fmt.Errorf("setVotingProxy: point %d is not a galaxy", point)
	}
	return packEcliptic("setVotingProxy", uint8(point), votingProxy)
}

func PackSpawn(point uint32, target common.Address) ([]byte, error) {
	return packEcliptic("spawn", point, target)
}

func PackTransferPoint(point uint32, target common.Address, reset bool) ([]byte, error) {
	return packEcliptic("transferPoint", point, target, reset)
}

func PackEscape(point, sponsor uint32) ([]byte, error) {
	return packEcliptic("escape", point, sponsor)
}

func PackCancelEscape(point uint32) ([]byte, error) {
	return packEcliptic("cancelEscape", point)
}

func PackAdopt(escapee uint32) ([]byte, error) {
	return packEcliptic("adopt", escapee)
}

func PackReject(escapee uint32) ([]byte, error) {
	return packEcliptic("reject", escapee)
}

func PackDetach(point uint32) ([]byte, error) {
	return packEcliptic("detach", point)
}

//
// Private
//

func packEcliptic(name string, args ...any) ([]byte, error) {
	c, err := NewEclipticContract()
	if err != nil {
		return nil, err
	}
	return c.Pack(name, args...)
}

func asBytes32(b []byte) ([32]byte, error) {
	var out [32]byte
	if len(b) != 32 {
		return out, fmt.Errorf("want 32 bytes, got %d", len(b))
	}
	copy(out[:], b)
	return out, nil
}
