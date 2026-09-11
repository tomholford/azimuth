package azimuth

import "github.com/ethereum/go-ethereum/common"

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
