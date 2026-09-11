package azimuth

import (
	"bytes"
	"embed"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

//
// Constants
//

const (
	// AzimuthAddress is azimuth.eth. It never changes.
	AzimuthAddress = "0x223c067F8CF28ae173EE5CafEa60cA44C335fecB"
	// EclipticAddress is ecliptic.eth on mainnet. Prefer Azimuth owner()
	// when a live client is available; this is the known fallback.
	EclipticAddress = "0x33EeCbf908478C10614626A9D304bfe18B78DD73"
)

//go:embed abi/azimuth.json abi/ecliptic.json
var abiFS embed.FS

var (
	azimuthABIOnce sync.Once
	azimuthABI     abi.ABI
	azimuthABIErr  error

	eclipticABIOnce sync.Once
	eclipticABI     abi.ABI
	eclipticABIErr  error
)

//
// Public
//

func AzimuthABI() (abi.ABI, error) {
	azimuthABIOnce.Do(func() {
		azimuthABI, azimuthABIErr = parseABI("abi/azimuth.json")
	})
	return azimuthABI, azimuthABIErr
}

func EclipticABI() (abi.ABI, error) {
	eclipticABIOnce.Do(func() {
		eclipticABI, eclipticABIErr = parseABI("abi/ecliptic.json")
	})
	return eclipticABI, eclipticABIErr
}

func AzimuthAddr() common.Address {
	return common.HexToAddress(AzimuthAddress)
}

func EclipticAddr() common.Address {
	return common.HexToAddress(EclipticAddress)
}

//
// Private
//

func parseABI(name string) (abi.ABI, error) {
	b, err := abiFS.ReadFile(name)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("read %s: %w", name, err)
	}
	parsed, err := abi.JSON(bytes.NewReader(b))
	if err != nil {
		return abi.ABI{}, fmt.Errorf("parse %s: %w", name, err)
	}
	return parsed, nil
}
