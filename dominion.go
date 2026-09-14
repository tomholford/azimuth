package azimuth

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

//
// Constants
//

const (
	DominionL1    = "l1"
	DominionL2    = "l2"
	DominionSpawn = "spawn"

	KeysPathEcliptic = "ecliptic"
	KeysPathRoller   = "roller"
)

//
// Public
//

func IsL1Dominion(d string) bool {
	return d == DominionL1
}

// DominionFromDeed infers L1/L2/spawn from Azimuth owner and spawn proxy.
// Owner at the deposit address means the point itself is on L2; spawn proxy
// at the deposit address means only spawn rights were deposited.
func DominionFromDeed(owner, spawnProxy common.Address) string {
	deposit := DepositAddr()
	if owner == deposit {
		return DominionL2
	}
	if spawnProxy == deposit {
		return DominionSpawn
	}
	return DominionL1
}

// KeysWritePath is where keys set submits: L1 and spawn stay on Ecliptic;
// L2 goes through the roller. Empty / unknown dominions fail closed.
func KeysWritePath(dominion string) (string, error) {
	switch dominion {
	case DominionL1, DominionSpawn:
		return KeysPathEcliptic, nil
	case DominionL2:
		return KeysPathRoller, nil
	case "":
		return "", fmt.Errorf("unknown dominion")
	default:
		return "", fmt.Errorf("unsupported dominion %s", dominion)
	}
}

// SpawnWritePath is where spawn, setSpawnProxy, and sponsor actions submit:
// L1 stays on Ecliptic; spawn and L2 go through the roller.
func SpawnWritePath(dominion string) (string, error) {
	switch dominion {
	case DominionL1:
		return KeysPathEcliptic, nil
	case DominionSpawn, DominionL2:
		return KeysPathRoller, nil
	case "":
		return "", fmt.Errorf("unknown dominion")
	default:
		return "", fmt.Errorf("unsupported dominion %s", dominion)
	}
}
