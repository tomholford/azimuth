package azimuth

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

//
// Types
//

// Contract is an ABI plus on-chain address.
type Contract struct {
	ABI     abi.ABI
	Address common.Address
}

//
// Public
//

func NewAzimuthContract() (*Contract, error) {
	parsed, err := AzimuthABI()
	if err != nil {
		return nil, err
	}
	return &Contract{ABI: parsed, Address: AzimuthAddr()}, nil
}

func NewEclipticContract() (*Contract, error) {
	parsed, err := EclipticABI()
	if err != nil {
		return nil, err
	}
	return &Contract{ABI: parsed, Address: EclipticAddr()}, nil
}

func NewClaimsContract() (*Contract, error) {
	parsed, err := ClaimsABI()
	if err != nil {
		return nil, err
	}
	return &Contract{ABI: parsed, Address: ClaimsAddr()}, nil
}

func (c *Contract) Pack(name string, args ...any) ([]byte, error) {
	data, err := c.ABI.Pack(name, args...)
	if err != nil {
		return nil, fmt.Errorf("pack %s: %w", name, err)
	}
	return data, nil
}

func (c *Contract) Call(ctx context.Context, client *ethclient.Client, name string, args ...any) ([]any, error) {
	data, err := c.Pack(name, args...)
	if err != nil {
		return nil, err
	}
	raw, err := client.CallContract(ctx, ethereum.CallMsg{To: &c.Address, Data: data}, nil)
	if err != nil {
		return nil, fmt.Errorf("eth_call %s: %w", name, err)
	}
	out, err := c.ABI.Unpack(name, raw)
	if err != nil {
		return nil, fmt.Errorf("unpack %s: %w", name, err)
	}
	return out, nil
}
