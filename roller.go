package azimuth

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

//
// Types
//

type Roller struct {
	apiURL url.URL
	http   *http.Client
}

type PointParams struct {
	Ship string `json:"ship"`
}

type AddressNonce struct {
	Address string `json:"address"`
	Nonce   int    `json:"nonce"`
}

type Ownership struct {
	Owner           AddressNonce `json:"owner"`
	ManagementProxy AddressNonce `json:"managementProxy"`
	SpawnProxy      AddressNonce `json:"spawnProxy"`
	TransferProxy   AddressNonce `json:"transferProxy"`
	VotingProxy     AddressNonce `json:"votingProxy"`
}

type NetworkKeys struct {
	Life  string `json:"life"`
	Suite string `json:"suite"`
	Auth  string `json:"auth"`
	Crypt string `json:"crypt"`
}

type Network struct {
	Keys    NetworkKeys `json:"keys"`
	Sponsor struct {
		Has bool `json:"has"`
		Who int  `json:"who"`
	} `json:"sponsor"`
	Rift string `json:"rift"`
}

type Result struct {
	Dominion  string    `json:"dominion"`
	Ownership Ownership `json:"ownership"`
	Network   Network   `json:"network"`
}

// L2From is the ship + proxy that issues a roller transaction.
type L2From struct {
	Ship  string `json:"ship"`
	Proxy string `json:"proxy"`
}

// ConfigureKeysData is the roller configureKeys payload (Bridge/azimuth-roll-rpc).
type ConfigureKeysData struct {
	Encrypt     string `json:"encrypt"`
	Auth        string `json:"auth"`
	CryptoSuite string `json:"cryptoSuite"`
	Breach      bool   `json:"breach"`
}

// AddressData is the roller payload for set*Proxy.
type AddressData struct {
	Address string `json:"address"`
}

// SpawnData is the roller spawn payload (child ship + recipient).
type SpawnData struct {
	Address string `json:"address"`
	Ship    string `json:"ship"`
}

// TransferPointData is the roller transferPoint payload.
type TransferPointData struct {
	Address string `json:"address"`
	Reset   bool   `json:"reset"`
}

// ShipData is the roller payload for escape/cancelEscape/adopt/reject/detach.
type ShipData struct {
	Ship string `json:"ship"`
}

//
// Constants
//

const (
	DefaultRollerURL = "https://roller.urbit.org/v1/roller"

	L2TxConfigureKeys      = "configureKeys"
	L2TxSetManagementProxy = "setManagementProxy"
	L2TxSetSpawnProxy      = "setSpawnProxy"
	L2TxSetTransferProxy   = "setTransferProxy"
	L2TxSetVotingProxy     = "setVotingProxy"
	L2TxSpawn              = "spawn"
	L2TxTransferPoint      = "transferPoint"
	L2TxEscape             = "escape"
	L2TxCancelEscape       = "cancelEscape"
	L2TxAdopt              = "adopt"
	L2TxReject             = "reject"
	L2TxDetach             = "detach"

	ProxyOwn      = "own"
	ProxyManage   = "manage"
	ProxySpawn    = "spawn"
	ProxyTransfer = "transfer"
	ProxyVote     = "vote"
)

var defaultAPIURL = url.URL{
	Scheme: "https",
	Host:   "roller.urbit.org",
	Path:   "/v1/roller",
}

//
// Public
//

func (r *Roller) GetPoint(ctx context.Context, point uint32) (*Result, error) {
	name, err := FormatPoint(point)
	if err != nil {
		return nil, err
	}
	var res Result
	if err := r.rpc(ctx, "getPoint", PointParams{Ship: name}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *Roller) GetShips(ctx context.Context, address string) ([]uint32, error) {
	var raw json.RawMessage
	if err := r.rpc(ctx, "getShips", map[string]string{"address": address}, &raw); err != nil {
		return nil, err
	}
	return parseShipIDs(raw)
}

// PrepareForSigning asks the roller for the EIP-191 message Frame/Metamask
// personal_sign. Matches Bridge api.prepareForSigning.
func (r *Roller) PrepareForSigning(ctx context.Context, nonce int, from L2From, tx string, data any) (string, error) {
	var hash string
	params := map[string]any{
		"nonce": nonce,
		"from":  from,
		"tx":    tx,
		"data":  data,
	}
	if err := r.rpc(ctx, "prepareForSigning", params, &hash); err != nil {
		return "", err
	}
	if hash == "" {
		return "", fmt.Errorf("roller prepareForSigning: empty hash")
	}
	return hash, nil
}

// ConfigureKeys submits a signed L2 configureKeys tx. Returns the roller tx hash.
func (r *Roller) ConfigureKeys(ctx context.Context, sig string, from L2From, address string, data ConfigureKeysData) (string, error) {
	return r.send(ctx, L2TxConfigureKeys, sig, from, address, data)
}

func (r *Roller) SetManagementProxy(ctx context.Context, sig string, from L2From, address string, data AddressData) (string, error) {
	return r.send(ctx, L2TxSetManagementProxy, sig, from, address, data)
}

func (r *Roller) SetSpawnProxy(ctx context.Context, sig string, from L2From, address string, data AddressData) (string, error) {
	return r.send(ctx, L2TxSetSpawnProxy, sig, from, address, data)
}

func (r *Roller) SetTransferProxy(ctx context.Context, sig string, from L2From, address string, data AddressData) (string, error) {
	return r.send(ctx, L2TxSetTransferProxy, sig, from, address, data)
}

func (r *Roller) SetVotingProxy(ctx context.Context, sig string, from L2From, address string, data AddressData) (string, error) {
	return r.send(ctx, L2TxSetVotingProxy, sig, from, address, data)
}

func (r *Roller) Spawn(ctx context.Context, sig string, from L2From, address string, data SpawnData) (string, error) {
	return r.send(ctx, L2TxSpawn, sig, from, address, data)
}

func (r *Roller) TransferPoint(ctx context.Context, sig string, from L2From, address string, data TransferPointData) (string, error) {
	return r.send(ctx, L2TxTransferPoint, sig, from, address, data)
}

func (r *Roller) Escape(ctx context.Context, sig string, from L2From, address string, data ShipData) (string, error) {
	return r.send(ctx, L2TxEscape, sig, from, address, data)
}

func (r *Roller) CancelEscape(ctx context.Context, sig string, from L2From, address string, data ShipData) (string, error) {
	return r.send(ctx, L2TxCancelEscape, sig, from, address, data)
}

func (r *Roller) Adopt(ctx context.Context, sig string, from L2From, address string, data ShipData) (string, error) {
	return r.send(ctx, L2TxAdopt, sig, from, address, data)
}

func (r *Roller) Reject(ctx context.Context, sig string, from L2From, address string, data ShipData) (string, error) {
	return r.send(ctx, L2TxReject, sig, from, address, data)
}

func (r *Roller) Detach(ctx context.Context, sig string, from L2From, address string, data ShipData) (string, error) {
	return r.send(ctx, L2TxDetach, sig, from, address, data)
}

func (r *Roller) GetTransactionStatus(ctx context.Context, hash string) (string, error) {
	var status string
	if err := r.rpc(ctx, "getTransactionStatus", map[string]string{"hash": hash}, &status); err != nil {
		return "", err
	}
	return status, nil
}

func NewSpawnData(point uint32, target common.Address) (SpawnData, error) {
	name, err := FormatPoint(point)
	if err != nil {
		return SpawnData{}, err
	}
	return SpawnData{Address: target.Hex(), Ship: name}, nil
}

func NewTransferPointData(target common.Address, reset bool) TransferPointData {
	return TransferPointData{Address: target.Hex(), Reset: reset}
}

func NewShipData(point uint32) (ShipData, error) {
	name, err := FormatPoint(point)
	if err != nil {
		return ShipData{}, err
	}
	return ShipData{Ship: name}, nil
}

func NewConfigureKeysData(encrypt, auth []byte, suite uint32, breach bool) (ConfigureKeysData, error) {
	if len(encrypt) != 32 {
		return ConfigureKeysData{}, fmt.Errorf("encrypt: want 32 bytes, got %d", len(encrypt))
	}
	if len(auth) != 32 {
		return ConfigureKeysData{}, fmt.Errorf("auth: want 32 bytes, got %d", len(auth))
	}
	return ConfigureKeysData{
		Encrypt:     "0x" + hex.EncodeToString(encrypt),
		Auth:        "0x" + hex.EncodeToString(auth),
		CryptoSuite: strconv.FormatUint(uint64(suite), 10),
		Breach:      breach,
	}, nil
}

func (r Result) ToPoint(index uint32) (Point, error) {
	name, err := FormatPoint(index)
	if err != nil {
		return Point{}, err
	}
	ed, crypt, err := r.convertHexToBytes()
	if err != nil {
		return Point{}, fmt.Errorf("failed to convert hex to bytes: %w", err)
	}
	suite, err := strconv.ParseUint(r.Network.Keys.Suite, 10, 32)
	if err != nil {
		return Point{}, fmt.Errorf("failed to parse decimal int from string %s: %w", r.Network.Keys.Suite, err)
	}
	revision, err := strconv.ParseUint(r.Network.Keys.Life, 10, 32)
	if err != nil {
		return Point{}, fmt.Errorf("failed to parse decimal int from string %s: %w", r.Network.Keys.Life, err)
	}
	var rift uint64
	if r.Network.Rift != "" {
		rift, err = strconv.ParseUint(r.Network.Rift, 10, 32)
		if err != nil {
			return Point{}, fmt.Errorf("failed to parse rift %s: %w", r.Network.Rift, err)
		}
	}
	if r.Network.Sponsor.Who < 0 {
		return Point{}, fmt.Errorf("negative sponsor %d", r.Network.Sponsor.Who)
	}
	return Point{
		Index:           index,
		Name:            name,
		Dominion:        r.Dominion,
		Owner:           r.Ownership.Owner.Proxy(),
		ManagementProxy: r.Ownership.ManagementProxy.Proxy(),
		SpawnProxy:      r.Ownership.SpawnProxy.Proxy(),
		TransferProxy:   r.Ownership.TransferProxy.Proxy(),
		VotingProxy:     r.Ownership.VotingProxy.Proxy(),
		CryptKey:        crypt,
		EdKey:           ed,
		Suite:           uint32(suite),
		Revision:        uint32(revision),
		Rift:            uint32(rift),
		HasSponsor:      r.Network.Sponsor.Has,
		Sponsor:         uint32(r.Network.Sponsor.Who),
	}, nil
}

// KeysProxy picks own vs manage from roller ownership for configureKeys.
func (r *Result) KeysProxy(addr string) (proxy string, nonce int, err error) {
	if !common.IsHexAddress(addr) {
		return "", 0, fmt.Errorf("invalid address %q", addr)
	}
	want := common.HexToAddress(addr)
	if want == (common.Address{}) {
		return "", 0, fmt.Errorf("zero address")
	}
	if owner := parseProxyAddr(r.Ownership.Owner.Address); owner == want {
		return ProxyOwn, r.Ownership.Owner.Nonce, nil
	}
	if mgr := parseProxyAddr(r.Ownership.ManagementProxy.Address); mgr == want {
		return ProxyManage, r.Ownership.ManagementProxy.Nonce, nil
	}
	return "", 0, fmt.Errorf("not owner or management proxy")
}

func (a AddressNonce) Proxy() Proxy {
	return Proxy{Address: parseProxyAddr(a.Address), Nonce: a.Nonce}
}

func (r Result) IsL1() bool {
	return IsL1Dominion(r.Dominion)
}

func NewRoller() *Roller {
	r, err := NewRollerURL("")
	if err != nil {
		return &Roller{apiURL: defaultAPIURL, http: defaultHTTP()}
	}
	return r
}

func NewRollerURL(raw string) (*Roller, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return &Roller{apiURL: defaultAPIURL, http: defaultHTTP()}, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("roller URL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("roller URL %q: need scheme and host", raw)
	}
	return &Roller{apiURL: *u, http: defaultHTTP()}, nil
}

//
// Private
//

func defaultHTTP() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func parseProxyAddr(s string) common.Address {
	if !common.IsHexAddress(s) {
		return common.Address{}
	}
	return common.HexToAddress(s)
}

func (r *Roller) send(ctx context.Context, method, sig string, from L2From, address string, data any) (string, error) {
	var hash string
	params := map[string]any{
		"sig":     sig,
		"from":    from,
		"address": address,
		"data":    data,
		"force":   false,
	}
	if err := r.rpc(ctx, method, params, &hash); err != nil {
		return "", err
	}
	if hash == "" {
		return "", fmt.Errorf("roller %s: empty hash", method)
	}
	return hash, nil
}

func (r *Roller) rpc(ctx context.Context, method string, params, result any) error {
	requestBody, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      "1337",
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.apiURL.String(), bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}
	var envelope struct {
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		Result json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	if envelope.Error != nil {
		return fmt.Errorf("roller %s: %s", method, envelope.Error.Message)
	}
	if err := json.Unmarshal(envelope.Result, result); err != nil {
		return fmt.Errorf("failed to decode %s result: %w", method, err)
	}
	return nil
}

func parseShipIDs(raw json.RawMessage) ([]uint32, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var nums []uint32
	if err := json.Unmarshal(raw, &nums); err == nil {
		return nums, nil
	}
	var strs []string
	if err := json.Unmarshal(raw, &strs); err != nil {
		return nil, fmt.Errorf("getShips: unexpected result %s", raw)
	}
	out := make([]uint32, 0, len(strs))
	for _, s := range strs {
		if n, err := strconv.ParseUint(s, 10, 32); err == nil {
			out = append(out, uint32(n))
			continue
		}
		idx, err := ParsePoint(s)
		if err != nil {
			return nil, fmt.Errorf("getShips: %q: %w", s, err)
		}
		out = append(out, idx)
	}
	return out, nil
}

func (r Result) convertHexToBytes() (authBytes, cryptBytes []byte, err error) {
	authBytes = make([]byte, 32)
	cryptBytes = make([]byte, 32)

	rawBytes, err := hex.DecodeString(strings.TrimPrefix(r.Network.Keys.Auth, "0x"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode auth hex string: %w", err)
	}
	copy(authBytes, rawBytes)

	rawBytes, err = hex.DecodeString(strings.TrimPrefix(r.Network.Keys.Crypt, "0x"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode crypt hex string: %w", err)
	}
	copy(cryptBytes, rawBytes)

	return authBytes, cryptBytes, nil
}
