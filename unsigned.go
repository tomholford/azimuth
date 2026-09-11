package azimuth

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

//
// Constants
//

const (
	LayerL1 = "l1"
	LayerL2 = "l2"
)

//
// Types
//

// Unsigned is calldata or a roller payload ready to sign.
// Signing stays with the caller: L1 uses eth_sendTransaction(To, Data);
// L2 uses personal_sign(Hash) then the matching roller submit RPC with that
// signature. This package does not implement a Signer.
type Unsigned struct {
	Layer string

	// L1
	To   common.Address
	Data []byte

	// L2
	From    L2From
	Nonce   int
	Tx      string
	Payload any
	Hash    string
}

// SubmitL2 posts a signed L2 payload to the roller. address is the Ethereum
// signer (the same from used when packing). Signing stays with the caller.
func (u *Unsigned) SubmitL2(ctx context.Context, r *Roller, sig, address string) (string, error) {
	if u == nil {
		return "", fmt.Errorf("nil unsigned")
	}
	if r == nil {
		return "", fmt.Errorf("nil roller")
	}
	if u.Layer != LayerL2 {
		return "", fmt.Errorf("submit L2: layer %q", u.Layer)
	}
	switch u.Tx {
	case L2TxConfigureKeys:
		data, ok := u.Payload.(ConfigureKeysData)
		if !ok {
			return "", fmt.Errorf("submit L2 %s: payload %T", u.Tx, u.Payload)
		}
		return r.ConfigureKeys(ctx, sig, u.From, address, data)
	case L2TxSetManagementProxy:
		data, ok := u.Payload.(AddressData)
		if !ok {
			return "", fmt.Errorf("submit L2 %s: payload %T", u.Tx, u.Payload)
		}
		return r.SetManagementProxy(ctx, sig, u.From, address, data)
	case L2TxSetSpawnProxy:
		data, ok := u.Payload.(AddressData)
		if !ok {
			return "", fmt.Errorf("submit L2 %s: payload %T", u.Tx, u.Payload)
		}
		return r.SetSpawnProxy(ctx, sig, u.From, address, data)
	case L2TxSetTransferProxy:
		data, ok := u.Payload.(AddressData)
		if !ok {
			return "", fmt.Errorf("submit L2 %s: payload %T", u.Tx, u.Payload)
		}
		return r.SetTransferProxy(ctx, sig, u.From, address, data)
	case L2TxSetVotingProxy:
		data, ok := u.Payload.(AddressData)
		if !ok {
			return "", fmt.Errorf("submit L2 %s: payload %T", u.Tx, u.Payload)
		}
		return r.SetVotingProxy(ctx, sig, u.From, address, data)
	case L2TxSpawn:
		data, ok := u.Payload.(SpawnData)
		if !ok {
			return "", fmt.Errorf("submit L2 %s: payload %T", u.Tx, u.Payload)
		}
		return r.Spawn(ctx, sig, u.From, address, data)
	case L2TxTransferPoint:
		data, ok := u.Payload.(TransferPointData)
		if !ok {
			return "", fmt.Errorf("submit L2 %s: payload %T", u.Tx, u.Payload)
		}
		return r.TransferPoint(ctx, sig, u.From, address, data)
	case L2TxEscape:
		data, ok := u.Payload.(ShipData)
		if !ok {
			return "", fmt.Errorf("submit L2 %s: payload %T", u.Tx, u.Payload)
		}
		return r.Escape(ctx, sig, u.From, address, data)
	case L2TxCancelEscape:
		data, ok := u.Payload.(ShipData)
		if !ok {
			return "", fmt.Errorf("submit L2 %s: payload %T", u.Tx, u.Payload)
		}
		return r.CancelEscape(ctx, sig, u.From, address, data)
	case L2TxAdopt:
		data, ok := u.Payload.(ShipData)
		if !ok {
			return "", fmt.Errorf("submit L2 %s: payload %T", u.Tx, u.Payload)
		}
		return r.Adopt(ctx, sig, u.From, address, data)
	case L2TxReject:
		data, ok := u.Payload.(ShipData)
		if !ok {
			return "", fmt.Errorf("submit L2 %s: payload %T", u.Tx, u.Payload)
		}
		return r.Reject(ctx, sig, u.From, address, data)
	case L2TxDetach:
		data, ok := u.Payload.(ShipData)
		if !ok {
			return "", fmt.Errorf("submit L2 %s: payload %T", u.Tx, u.Payload)
		}
		return r.Detach(ctx, sig, u.From, address, data)
	default:
		return "", fmt.Errorf("submit L2: unknown tx %q", u.Tx)
	}
}
