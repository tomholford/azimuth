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
	// ClaimsAddress is the mainnet Claims contract. Prefer Ecliptic claims()
	// when a live client is available; this is the known fallback.
	ClaimsAddress = "0xe7e7f69b34D7d9Bd8d61Fb22C33b22708947971A"
	// DepositAddress is Ecliptic.depositAddress — ships sent here are on L2.
	// Prefer Ecliptic depositAddress() when a live client is available; this
	// is the known fallback (the constant never changes).
	DepositAddress = "0x1111111111111111111111111111111111111111"
)

//go:embed abi/azimuth.json abi/ecliptic.json abi/claims.json
var abiFS embed.FS

var (
	azimuthABIOnce sync.Once
	azimuthABI     abi.ABI
	azimuthABIErr  error

	eclipticABIOnce sync.Once
	eclipticABI     abi.ABI
	eclipticABIErr  error

	claimsABIOnce sync.Once
	claimsABI     abi.ABI
	claimsABIErr  error
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

func ClaimsABI() (abi.ABI, error) {
	claimsABIOnce.Do(func() {
		claimsABI, claimsABIErr = parseABI("abi/claims.json")
	})
	return claimsABI, claimsABIErr
}

func AzimuthAddr() common.Address {
	return common.HexToAddress(AzimuthAddress)
}

func EclipticAddr() common.Address {
	return common.HexToAddress(EclipticAddress)
}

func ClaimsAddr() common.Address {
	return common.HexToAddress(ClaimsAddress)
}

func DepositAddr() common.Address {
	return common.HexToAddress(DepositAddress)
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
