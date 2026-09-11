package azimuth

import "fmt"

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
