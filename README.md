# azimuth

Go client for Urbit [Azimuth](https://github.com/urbit/azimuth) / Ecliptic (L1)
and the L2 [roller](https://roller.urbit.org/v1/roller).

Reads point state, packs unsigned Ecliptic calls, and talks to the roller.
Signing stays with the caller (wallet, keystore, Frame, etc).

```
go get github.com/tomholford/azimuth
```

Requires Go 1.24+.

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"

	"github.com/tomholford/azimuth"
)

func main() {
	c, err := azimuth.Dial("https://ethereum-rpc.publicnode.com", "")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()
	p, err := c.GetPoint(ctx, 0) // ~zod
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(p.Name, p.Dominion, p.Owner.Address.Hex())

	from := p.Owner.Address
	u, err := c.ConfigureKeys(ctx, p.Index, p.CryptKey, p.EdKey, azimuth.CryptoSuiteVersion, false, from)
	if err != nil {
		log.Fatal(err)
	}
	switch u.Layer {
	case azimuth.LayerL1:
		// eth_sendTransaction({to: u.To, data: u.Data})
		fmt.Printf("L1 %s %x\n", u.To.Hex(), u.Data)
	case azimuth.LayerL2:
		// personal_sign(u.Hash), then roller submit with that signature
		fmt.Printf("L2 %s nonce=%d hash=%s\n", u.Tx, u.Nonce, u.Hash)
	}
}
```

`Dial` takes an Ethereum JSON-RPC URL and an optional roller URL (empty uses
`https://roller.urbit.org/v1/roller`). Writes return [`Unsigned`](unsigned.go):
this package does not sign or submit.

`@p` helpers: `ParsePoint("~zod")`, `FormatPoint(0)`, `Clan`, `Prefix`.

### Writes

`Client` reads dominion from the roller, then packs L1 calldata or an L2
`prepareForSigning` payload.

| Method | L1 / spawn dominion | L2 |
| --- | --- | --- |
| `ConfigureKeys`, `SetManagementProxy`, `SetTransferProxy`, `SetVotingProxy`, `TransferPoint` | Ecliptic | roller |
| `SetSpawnProxy`, `Spawn`, `Escape`, `CancelEscape`, `Adopt`, `Reject`, `Detach` | Ecliptic if dominion is `l1`; roller if `spawn` or `l2` | roller |

`Spawn` is issued by the child's prefix. `Adopt` / `Reject` / `Detach` take the
acting sponsor as the first point argument.

Direct ABI packers (`PackConfigureKeys`, `PackSpawn`, `PackTransferPoint`,
`PackEscape`, …) if you already know the layer.

### Contracts

- Azimuth (data): `0x223c067F8CF28ae173EE5CafEa60cA44C335fecB` — never changes
- Ecliptic (logic): `Azimuth.owner()`, falling back to
  `0x33EeCbf908478C10614626A9D304bfe18B78DD73`

## Dev

```sh
./ops/gauntlet.sh   # fmt → lint → test → build
./ops/fmt.sh
./ops/lint.sh
./ops/test.sh
./ops/build.sh
```

Requires [golangci-lint](https://golangci-lint.run/) v2. GitHub Actions runs the
gauntlet on pull requests and on merge to `master`.
