package azimuth

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

//
// Types
//

// Client talks to an Ethereum RPC and the L2 roller. Writes return Unsigned
// payloads; the caller signs and submits.
type Client struct {
	eth    *ethclient.Client
	roller *Roller
}

//
// Public
//

func Dial(ethRPC, rollerURL string) (*Client, error) {
	ethRPC = strings.TrimSpace(ethRPC)
	if ethRPC == "" {
		return nil, fmt.Errorf("empty eth RPC URL")
	}
	eth, err := ethclient.Dial(ethRPC)
	if err != nil {
		return nil, fmt.Errorf("eth RPC: %w", err)
	}
	roller, err := NewRollerURL(rollerURL)
	if err != nil {
		eth.Close()
		return nil, err
	}
	return &Client{eth: eth, roller: roller}, nil
}

func (c *Client) Close() {
	if c != nil && c.eth != nil {
		c.eth.Close()
	}
}

func (c *Client) Roller() *Roller {
	if c == nil {
		return nil
	}
	return c.roller
}

// GetPoint prefers the roller (L1+L2). On roller failure it falls back to
// Azimuth points() and rights(), inferring dominion from the deposit address.
func (c *Client) GetPoint(ctx context.Context, point uint32) (*Point, error) {
	var rollerErr error
	if c.roller != nil {
		res, err := c.roller.GetPoint(ctx, point)
		if err == nil {
			var p Point
			p, err = res.ToPoint(point)
			if err == nil {
				return &p, nil
			}
		}
		rollerErr = err
	}
	p, err := c.getPointL1(ctx, point)
	if err != nil {
		if rollerErr != nil {
			return nil, fmt.Errorf("roller: %w; l1: %w", rollerErr, err)
		}
		return nil, err
	}
	return p, nil
}

func (c *Client) ConfigureKeys(
	ctx context.Context,
	point uint32,
	encrypt, auth []byte,
	suite uint32,
	discontinuous bool,
	from common.Address,
) (*Unsigned, error) {
	path, res, err := c.route(ctx, point)
	if err != nil {
		return nil, err
	}
	if path == KeysPathEcliptic {
		var data []byte
		data, err = PackConfigureKeys(point, encrypt, auth, suite, discontinuous)
		if err != nil {
			return nil, err
		}
		return c.unsignedL1(ctx, data), nil
	}
	var payload ConfigureKeysData
	payload, err = NewConfigureKeysData(encrypt, auth, suite, discontinuous)
	if err != nil {
		return nil, err
	}
	return c.unsignedL2(ctx, res, point, from, L2TxConfigureKeys, payload, (*Result).KeysProxy)
}

func (c *Client) SetManagementProxy(ctx context.Context, point uint32, manager, from common.Address) (*Unsigned, error) {
	return c.setProxy(ctx, point, manager, from, L2TxSetManagementProxy, PackSetManagementProxy, (*Result).KeysProxy)
}

func (c *Client) SetSpawnProxy(ctx context.Context, point uint32, spawnProxy, from common.Address) (*Unsigned, error) {
	path, res, err := c.routePath(ctx, point, SpawnWritePath)
	if err != nil {
		return nil, err
	}
	if path == KeysPathEcliptic {
		var data []byte
		data, err = PackSetSpawnProxy(point, spawnProxy)
		if err != nil {
			return nil, err
		}
		return c.unsignedL1(ctx, data), nil
	}
	return c.unsignedL2(ctx, res, point, from, L2TxSetSpawnProxy, AddressData{Address: spawnProxy.Hex()}, spawnProxyRole)
}

func (c *Client) Spawn(ctx context.Context, point uint32, target, from common.Address) (*Unsigned, error) {
	prefix, err := Prefix(point)
	if err != nil {
		return nil, err
	}
	path, res, err := c.routePath(ctx, prefix, SpawnWritePath)
	if err != nil {
		return nil, err
	}
	if path == KeysPathEcliptic {
		var data []byte
		data, err = PackSpawn(point, target)
		if err != nil {
			return nil, err
		}
		return c.unsignedL1(ctx, data), nil
	}
	var payload SpawnData
	payload, err = NewSpawnData(point, target)
	if err != nil {
		return nil, err
	}
	return c.unsignedL2(ctx, res, prefix, from, L2TxSpawn, payload, spawnProxyRole)
}

func (c *Client) TransferPoint(
	ctx context.Context,
	point uint32,
	target common.Address,
	reset bool,
	from common.Address,
) (*Unsigned, error) {
	path, res, err := c.route(ctx, point)
	if err != nil {
		return nil, err
	}
	if path == KeysPathEcliptic {
		var data []byte
		data, err = PackTransferPoint(point, target, reset)
		if err != nil {
			return nil, err
		}
		return c.unsignedL1(ctx, data), nil
	}
	return c.unsignedL2(ctx, res, point, from, L2TxTransferPoint, NewTransferPointData(target, reset), transferProxyRole)
}

func (c *Client) Escape(ctx context.Context, point, sponsor uint32, from common.Address) (*Unsigned, error) {
	path, res, err := c.routePath(ctx, point, SpawnWritePath)
	if err != nil {
		return nil, err
	}
	if path == KeysPathEcliptic {
		var data []byte
		data, err = PackEscape(point, sponsor)
		if err != nil {
			return nil, err
		}
		return c.unsignedL1(ctx, data), nil
	}
	var payload ShipData
	payload, err = NewShipData(sponsor)
	if err != nil {
		return nil, err
	}
	return c.unsignedL2(ctx, res, point, from, L2TxEscape, payload, (*Result).KeysProxy)
}

func (c *Client) CancelEscape(ctx context.Context, point uint32, from common.Address) (*Unsigned, error) {
	return c.sponsorTx(ctx, point, point, from, L2TxCancelEscape, PackCancelEscape, (*Result).KeysProxy)
}

func (c *Client) Adopt(ctx context.Context, sponsor, escapee uint32, from common.Address) (*Unsigned, error) {
	return c.sponsorTx(ctx, sponsor, escapee, from, L2TxAdopt, PackAdopt, (*Result).KeysProxy)
}

func (c *Client) Reject(ctx context.Context, sponsor, escapee uint32, from common.Address) (*Unsigned, error) {
	return c.sponsorTx(ctx, sponsor, escapee, from, L2TxReject, PackReject, (*Result).KeysProxy)
}

func (c *Client) Detach(ctx context.Context, sponsor, point uint32, from common.Address) (*Unsigned, error) {
	return c.sponsorTx(ctx, sponsor, point, from, L2TxDetach, PackDetach, (*Result).KeysProxy)
}

func (c *Client) SetTransferProxy(ctx context.Context, point uint32, transferProxy, from common.Address) (*Unsigned, error) {
	return c.setProxy(ctx, point, transferProxy, from, L2TxSetTransferProxy, PackSetTransferProxy, transferProxyRole)
}

func (c *Client) SetVotingProxy(ctx context.Context, point uint32, votingProxy, from common.Address) (*Unsigned, error) {
	return c.setProxy(ctx, point, votingProxy, from, L2TxSetVotingProxy, PackSetVotingProxy, votingProxyRole)
}

// AddClaim packs an L1 Claims.addClaim. Claims is L1-only; the roller has no claim txs.
func (c *Client) AddClaim(ctx context.Context, point uint32, protocol, claim string, dossier []byte) (*Unsigned, error) {
	data, err := PackAddClaim(point, protocol, claim, dossier)
	if err != nil {
		return nil, err
	}
	return c.unsignedClaims(ctx, data), nil
}

func (c *Client) RemoveClaim(ctx context.Context, point uint32, protocol, claim string) (*Unsigned, error) {
	data, err := PackRemoveClaim(point, protocol, claim)
	if err != nil {
		return nil, err
	}
	return c.unsignedClaims(ctx, data), nil
}

func (c *Client) ClearClaims(ctx context.Context, point uint32) (*Unsigned, error) {
	data, err := PackClearClaims(point)
	if err != nil {
		return nil, err
	}
	return c.unsignedClaims(ctx, data), nil
}

func (c *Client) GetClaim(ctx context.Context, point uint32, index uint8) (*Claim, error) {
	if c == nil || c.eth == nil {
		return nil, fmt.Errorf("no eth client")
	}
	return GetClaim(ctx, c.eth, point, index)
}

func (c *Client) GetClaims(ctx context.Context, point uint32) ([]Claim, error) {
	if c == nil || c.eth == nil {
		return nil, fmt.Errorf("no eth client")
	}
	return GetClaims(ctx, c.eth, point)
}

// SubmitL2 posts a signed L2 Unsigned to this client's roller.
func (c *Client) SubmitL2(ctx context.Context, u *Unsigned, sig string, address common.Address) (string, error) {
	if c == nil || c.roller == nil {
		return "", fmt.Errorf("no roller")
	}
	return u.SubmitL2(ctx, c.roller, sig, address.Hex())
}

//
// Private
//

func (c *Client) route(ctx context.Context, point uint32) (string, *Result, error) {
	return c.routePath(ctx, point, KeysWritePath)
}

func (c *Client) routePath(ctx context.Context, point uint32, pathFn func(string) (string, error)) (string, *Result, error) {
	if c.roller == nil {
		return "", nil, fmt.Errorf("no roller")
	}
	res, err := c.roller.GetPoint(ctx, point)
	if err != nil {
		return "", nil, err
	}
	var path string
	path, err = pathFn(res.Dominion)
	if err != nil {
		return "", nil, err
	}
	return path, res, nil
}

func (c *Client) sponsorTx(
	ctx context.Context,
	fromPoint, dataPoint uint32,
	from common.Address,
	tx string,
	pack func(uint32) ([]byte, error),
	pick func(*Result, string) (string, int, error),
) (*Unsigned, error) {
	path, res, err := c.routePath(ctx, fromPoint, SpawnWritePath)
	if err != nil {
		return nil, err
	}
	if path == KeysPathEcliptic {
		var data []byte
		data, err = pack(dataPoint)
		if err != nil {
			return nil, err
		}
		return c.unsignedL1(ctx, data), nil
	}
	var payload ShipData
	payload, err = NewShipData(dataPoint)
	if err != nil {
		return nil, err
	}
	return c.unsignedL2(ctx, res, fromPoint, from, tx, payload, pick)
}

func (c *Client) setProxy(
	ctx context.Context,
	point uint32,
	proxyAddr, from common.Address,
	tx string,
	pack func(uint32, common.Address) ([]byte, error),
	pick func(*Result, string) (string, int, error),
) (*Unsigned, error) {
	path, res, err := c.route(ctx, point)
	if err != nil {
		return nil, err
	}
	if path == KeysPathEcliptic {
		var data []byte
		data, err = pack(point, proxyAddr)
		if err != nil {
			return nil, err
		}
		return c.unsignedL1(ctx, data), nil
	}
	return c.unsignedL2(ctx, res, point, from, tx, AddressData{Address: proxyAddr.Hex()}, pick)
}

func (c *Client) unsignedL1(ctx context.Context, data []byte) *Unsigned {
	return &Unsigned{Layer: LayerL1, To: c.eclipticAddr(ctx), Data: data}
}

func (c *Client) unsignedClaims(ctx context.Context, data []byte) *Unsigned {
	return &Unsigned{Layer: LayerL1, To: c.claimsAddr(ctx), Data: data}
}

func (c *Client) unsignedL2(
	ctx context.Context,
	res *Result,
	point uint32,
	from common.Address,
	tx string,
	payload any,
	pick func(*Result, string) (string, int, error),
) (*Unsigned, error) {
	proxy, nonce, err := pick(res, from.Hex())
	if err != nil {
		return nil, err
	}
	var name, hash string
	name, err = FormatPoint(point)
	if err != nil {
		return nil, err
	}
	l2from := L2From{Ship: name, Proxy: proxy}
	hash, err = c.roller.PrepareForSigning(ctx, nonce, l2from, tx, payload)
	if err != nil {
		return nil, err
	}
	return &Unsigned{
		Layer:   LayerL2,
		From:    l2from,
		Nonce:   nonce,
		Tx:      tx,
		Payload: payload,
		Hash:    hash,
	}, nil
}

func (c *Client) eclipticAddr(ctx context.Context) common.Address {
	if c.eth != nil {
		owner, err := Owner(ctx, c.eth)
		if err == nil && owner != (common.Address{}) {
			return owner
		}
	}
	return EclipticAddr()
}

func (c *Client) claimsAddr(ctx context.Context) common.Address {
	if c.eth == nil {
		return ClaimsAddr()
	}
	ec, err := NewEclipticContract()
	if err != nil {
		return ClaimsAddr()
	}
	ec.Address = c.eclipticAddr(ctx)
	out, err := ec.Call(ctx, c.eth, "claims")
	if err != nil || len(out) != 1 {
		return ClaimsAddr()
	}
	addr, err := unpackAddress(out[0])
	if err != nil || addr == (common.Address{}) {
		return ClaimsAddr()
	}
	return addr
}

func (c *Client) getPointL1(ctx context.Context, point uint32) (*Point, error) {
	if c.eth == nil {
		return nil, fmt.Errorf("no eth client")
	}
	name, err := FormatPoint(point)
	if err != nil {
		return nil, err
	}
	var data *PointData
	data, err = GetPointData(ctx, c.eth, point)
	if err != nil {
		return nil, err
	}
	var deed *Deed
	deed, err = GetRights(ctx, c.eth, point)
	if err != nil {
		return nil, err
	}
	return &Point{
		Index:             point,
		Name:              name,
		Dominion:          DominionFromDeed(deed.Owner, deed.SpawnProxy),
		Owner:             Proxy{Address: deed.Owner},
		ManagementProxy:   Proxy{Address: deed.ManagementProxy},
		SpawnProxy:        Proxy{Address: deed.SpawnProxy},
		TransferProxy:     Proxy{Address: deed.TransferProxy},
		VotingProxy:       Proxy{Address: deed.VotingProxy},
		CryptKey:          data.CryptKey,
		EdKey:             data.EdKey,
		Suite:             data.Suite,
		Revision:          data.Revision,
		Rift:              data.Rift,
		HasSponsor:        data.HasSponsor,
		Sponsor:           data.Sponsor,
		Active:            data.Active,
		EscapeRequested:   data.EscapeRequested,
		EscapeRequestedTo: data.EscapeRequestedTo,
	}, nil
}

func spawnProxyRole(res *Result, addr string) (string, int, error) {
	return matchProxy(addr,
		role{ProxyOwn, res.Ownership.Owner},
		role{ProxySpawn, res.Ownership.SpawnProxy},
	)
}

func transferProxyRole(res *Result, addr string) (string, int, error) {
	return matchProxy(addr,
		role{ProxyOwn, res.Ownership.Owner},
		role{ProxyTransfer, res.Ownership.TransferProxy},
	)
}

func votingProxyRole(res *Result, addr string) (string, int, error) {
	return matchProxy(addr,
		role{ProxyOwn, res.Ownership.Owner},
		role{ProxyVote, res.Ownership.VotingProxy},
	)
}

type role struct {
	proxy string
	an    AddressNonce
}

func matchProxy(addr string, roles ...role) (string, int, error) {
	if !common.IsHexAddress(addr) {
		return "", 0, fmt.Errorf("invalid address %q", addr)
	}
	want := common.HexToAddress(addr)
	if want == (common.Address{}) {
		return "", 0, fmt.Errorf("zero address")
	}
	for _, r := range roles {
		if parseProxyAddr(r.an.Address) == want {
			return r.proxy, r.an.Nonce, nil
		}
	}
	return "", 0, fmt.Errorf("address is not a permitted proxy")
}
