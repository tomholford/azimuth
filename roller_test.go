package azimuth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseShipIDs(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []uint32
		wantErr bool
	}{
		{name: "uints", raw: `[0,256,69]`, want: []uint32{0, 256, 69}},
		{name: "strings", raw: `["0","256"]`, want: []uint32{0, 256}},
		{name: "patp", raw: `["~zod","~pet"]`, want: []uint32{0, 69}},
		{name: "null", raw: `null`},
		{name: "empty", raw: `[]`, want: []uint32{}},
		{name: "bad", raw: `{"nope":true}`, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseShipIDs(json.RawMessage(tc.raw))
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("len %d want %d (%v)", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("idx %d: %d want %d", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestGetShipsMock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		if req.Method != "getShips" {
			t.Errorf("method %s", req.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1337","result":[69,256]}`))
	}))
	defer srv.Close()

	r := mustRoller(t, srv.URL)
	got, err := r.GetShips(context.Background(), "0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != 69 || got[1] != 256 {
		t.Fatalf("got %v", got)
	}
}

func TestGetPointMock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string      `json:"method"`
			Params PointParams `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		if req.Method != "getPoint" || req.Params.Ship != "~pet" {
			t.Errorf("got %+v", req)
		}
		res := Result{Dominion: DominionL2}
		res.Ownership.Owner.Address = "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
		res.Ownership.Owner.Nonce = 4
		res.Network.Keys.Life = "3"
		res.Network.Keys.Suite = "1"
		res.Network.Keys.Auth = "0x" + strings.Repeat("ab", 32)
		res.Network.Keys.Crypt = "0x" + strings.Repeat("cd", 32)
		res.Network.Sponsor.Has = true
		res.Network.Sponsor.Who = 69
		res.Network.Rift = "2"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": "1337", "result": res})
	}))
	defer srv.Close()

	r := mustRoller(t, srv.URL)
	res, err := r.GetPoint(context.Background(), 69)
	if err != nil {
		t.Fatal(err)
	}
	p, err := res.ToPoint(69)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "~pet" || p.Dominion != DominionL2 || p.Revision != 3 || p.Rift != 2 || p.Sponsor != 69 {
		t.Fatalf("%+v", p)
	}
	if p.Owner.Nonce != 4 {
		t.Fatalf("owner %+v", p.Owner)
	}
}

func TestResultIsL1(t *testing.T) {
	if !(Result{Dominion: DominionL1}).IsL1() {
		t.Fatal("l1")
	}
	if (Result{Dominion: DominionL2}).IsL1() {
		t.Fatal("l2")
	}
}

func TestNewRollerURL(t *testing.T) {
	r, err := NewRollerURL("")
	if err != nil {
		t.Fatal(err)
	}
	if r.apiURL != defaultAPIURL {
		t.Fatalf("empty: %s", r.apiURL.String())
	}
	r, err = NewRollerURL("http://127.0.0.1:8080/v1/roller")
	if err != nil {
		t.Fatal(err)
	}
	if r.apiURL.Host != "127.0.0.1:8080" || r.apiURL.Path != "/v1/roller" {
		t.Fatalf("got %s", r.apiURL.String())
	}
	if _, err := NewRollerURL("not-a-url"); err == nil {
		t.Fatal("expected error")
	}
}

func TestPrepareForSigningMock(t *testing.T) {
	crypt := bytes.Repeat([]byte{0x11}, 32)
	auth := bytes.Repeat([]byte{0x22}, 32)
	data, err := NewConfigureKeysData(crypt, auth, CryptoSuiteVersion, true)
	if err != nil {
		t.Fatal(err)
	}
	from := L2From{Ship: "~panted-noshes", Proxy: ProxyOwn}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string            `json:"method"`
			Params prepareSignParams `json:"params"`
		}
		if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
			t.Errorf("decode: %v", decErr)
			return
		}
		if req.Method != "prepareForSigning" {
			t.Errorf("method %s", req.Method)
		}
		if req.Params.Nonce != 3 || req.Params.Tx != L2TxConfigureKeys {
			t.Errorf("nonce/tx %+v", req.Params)
		}
		if req.Params.From != from {
			t.Errorf("from %+v", req.Params.From)
		}
		if req.Params.Data != data {
			t.Errorf("data %+v", req.Params.Data)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1337","result":"0xcafe"}`))
	}))
	defer srv.Close()

	roller := mustRoller(t, srv.URL)
	got, err := roller.PrepareForSigning(context.Background(), 3, from, L2TxConfigureKeys, data)
	if err != nil {
		t.Fatal(err)
	}
	if got != "0xcafe" {
		t.Fatalf("hash %q", got)
	}
}

func TestRollerSubmitMock(t *testing.T) {
	from := L2From{Ship: "~zod", Proxy: ProxyOwn}
	addr := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	sig := "0x" + strings.Repeat("ab", 65)
	proxyAddr := "0xF7306c5db0C1880FB2ed9c3972ad3e1A94999196"
	keys := ConfigureKeysData{
		Encrypt:     "0x" + strings.Repeat("11", 32),
		Auth:        "0x" + strings.Repeat("22", 32),
		CryptoSuite: "1",
	}
	proxy := AddressData{Address: proxyAddr}
	spawn := SpawnData{Address: proxyAddr, Ship: "~pet"}
	transfer := TransferPointData{Address: proxyAddr, Reset: true}
	ship := ShipData{Ship: "~wicdev-wisryt"}

	tests := []struct {
		method string
		data   any
		call   func(*Roller) (string, error)
	}{
		{L2TxConfigureKeys, keys, func(r *Roller) (string, error) {
			return r.ConfigureKeys(context.Background(), sig, from, addr, keys)
		}},
		{L2TxSetManagementProxy, proxy, func(r *Roller) (string, error) {
			return r.SetManagementProxy(context.Background(), sig, from, addr, proxy)
		}},
		{L2TxSetSpawnProxy, proxy, func(r *Roller) (string, error) {
			return r.SetSpawnProxy(context.Background(), sig, from, addr, proxy)
		}},
		{L2TxSetTransferProxy, proxy, func(r *Roller) (string, error) {
			return r.SetTransferProxy(context.Background(), sig, from, addr, proxy)
		}},
		{L2TxSetVotingProxy, proxy, func(r *Roller) (string, error) {
			return r.SetVotingProxy(context.Background(), sig, from, addr, proxy)
		}},
		{L2TxSpawn, spawn, func(r *Roller) (string, error) {
			return r.Spawn(context.Background(), sig, from, addr, spawn)
		}},
		{L2TxTransferPoint, transfer, func(r *Roller) (string, error) {
			return r.TransferPoint(context.Background(), sig, from, addr, transfer)
		}},
		{L2TxEscape, ship, func(r *Roller) (string, error) {
			return r.Escape(context.Background(), sig, from, addr, ship)
		}},
		{L2TxCancelEscape, ship, func(r *Roller) (string, error) {
			return r.CancelEscape(context.Background(), sig, from, addr, ship)
		}},
		{L2TxAdopt, ship, func(r *Roller) (string, error) {
			return r.Adopt(context.Background(), sig, from, addr, ship)
		}},
		{L2TxReject, ship, func(r *Roller) (string, error) {
			return r.Reject(context.Background(), sig, from, addr, ship)
		}},
		{L2TxDetach, ship, func(r *Roller) (string, error) {
			return r.Detach(context.Background(), sig, from, addr, ship)
		}},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			wantData, err := json.Marshal(tc.data)
			if err != nil {
				t.Fatal(err)
			}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req struct {
					Method string `json:"method"`
					Params struct {
						Sig     string          `json:"sig"`
						From    L2From          `json:"from"`
						Address string          `json:"address"`
						Data    json.RawMessage `json:"data"`
						Force   bool            `json:"force"`
					} `json:"params"`
				}
				if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
					t.Errorf("decode: %v", decErr)
					return
				}
				if req.Method != tc.method {
					t.Errorf("method %s want %s", req.Method, tc.method)
				}
				if req.Params.Sig != sig || req.Params.Address != addr || req.Params.Force {
					t.Errorf("params %+v", req.Params)
				}
				if req.Params.From != from {
					t.Errorf("from %+v", req.Params.From)
				}
				gotData := bytes.TrimSpace(req.Params.Data)
				if !bytes.Equal(gotData, wantData) {
					t.Errorf("data %s want %s", gotData, wantData)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1337","result":"0xbeef"}`))
			}))
			defer srv.Close()

			got, err := tc.call(mustRoller(t, srv.URL))
			if err != nil {
				t.Fatal(err)
			}
			if got != "0xbeef" {
				t.Fatalf("hash %q", got)
			}
		})
	}
}

func TestGetTransactionStatusMock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string            `json:"method"`
			Params map[string]string `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		if req.Method != "getTransactionStatus" || req.Params["hash"] != "0xbeef" {
			t.Errorf("got %+v", req)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1337","result":"pending"}`))
	}))
	defer srv.Close()

	roller := mustRoller(t, srv.URL)
	got, err := roller.GetTransactionStatus(context.Background(), "0xbeef")
	if err != nil {
		t.Fatal(err)
	}
	if got != "pending" {
		t.Fatalf("status %q", got)
	}
}

func TestRollerRPCError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1337","error":{"code":-32002,"message":"Max tx quota exceeded"}}`))
	}))
	defer srv.Close()

	roller := mustRoller(t, srv.URL)
	_, err := roller.ConfigureKeys(context.Background(), "0x00", L2From{Ship: "~zod", Proxy: ProxyOwn}, "0x00", ConfigureKeysData{})
	if err == nil || !strings.Contains(err.Error(), "Max tx quota exceeded") {
		t.Fatalf("err=%v", err)
	}
}

func TestKeysProxy(t *testing.T) {
	owner := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	mgr := "0x1111111111111111111111111111111111111111"
	res := Result{}
	res.Ownership.Owner.Address = owner
	res.Ownership.Owner.Nonce = 4
	res.Ownership.ManagementProxy.Address = mgr
	res.Ownership.ManagementProxy.Nonce = 1

	proxy, nonce, err := res.KeysProxy(strings.ToLower(owner))
	if err != nil || proxy != ProxyOwn || nonce != 4 {
		t.Fatalf("owner: %s %d %v", proxy, nonce, err)
	}
	proxy, nonce, err = res.KeysProxy(mgr)
	if err != nil || proxy != ProxyManage || nonce != 1 {
		t.Fatalf("manage: %s %d %v", proxy, nonce, err)
	}
	if _, _, err := res.KeysProxy("0x2222222222222222222222222222222222222222"); err == nil {
		t.Fatal("expected not owner")
	}
	if _, _, err := res.KeysProxy("nope"); err == nil {
		t.Fatal("expected invalid")
	}
	if _, _, err := res.KeysProxy("0x0000000000000000000000000000000000000000"); err == nil {
		t.Fatal("expected zero")
	}
}

func TestNewConfigureKeysData(t *testing.T) {
	crypt := bytes.Repeat([]byte{0xaa}, 32)
	auth := bytes.Repeat([]byte{0xbb}, 32)
	data, err := NewConfigureKeysData(crypt, auth, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	if data.Encrypt != "0x"+strings.Repeat("aa", 32) || data.Auth != "0x"+strings.Repeat("bb", 32) {
		t.Fatalf("%+v", data)
	}
	if data.CryptoSuite != "1" || data.Breach {
		t.Fatalf("%+v", data)
	}
	if _, err := NewConfigureKeysData([]byte{1}, auth, 1, false); err == nil {
		t.Fatal("short encrypt")
	}
}

func TestResultToPoint(t *testing.T) {
	validAuth := "0x" + strings.Repeat("ab", 32)
	validCrypt := "0x" + strings.Repeat("cd", 32)
	validResult := func() Result {
		return Result{
			Dominion: DominionL2,
			Network: Network{
				Keys: NetworkKeys{
					Life:  "3",
					Suite: "1",
					Auth:  validAuth,
					Crypt: validCrypt,
				},
				Sponsor: struct {
					Has bool `json:"has"`
					Who int  `json:"who"`
				}{Has: true, Who: 42},
				Rift: "7",
			},
		}
	}

	p, err := validResult().ToPoint(0)
	if err != nil {
		t.Fatal(err)
	}
	if p.Index != 0 || p.Name != "~zod" || p.Dominion != DominionL2 {
		t.Fatalf("%+v", p)
	}
	if p.Suite != 1 || p.Revision != 3 || p.Rift != 7 || !p.HasSponsor || p.Sponsor != 42 {
		t.Fatalf("%+v", p)
	}
	if len(p.EdKey) != 32 || len(p.CryptKey) != 32 {
		t.Fatalf("key lens %d %d", len(p.EdKey), len(p.CryptKey))
	}

	bad := validResult()
	bad.Network.Keys.Auth = "0xzzzz"
	if _, err := bad.ToPoint(0); err == nil {
		t.Fatal("bad auth")
	}
	bad = validResult()
	bad.Network.Keys.Suite = "abc"
	if _, err := bad.ToPoint(0); err == nil {
		t.Fatal("bad suite")
	}
}

type prepareSignParams struct {
	Nonce int               `json:"nonce"`
	Tx    string            `json:"tx"`
	From  L2From            `json:"from"`
	Data  ConfigureKeysData `json:"data"`
}

func mustRoller(t *testing.T, raw string) *Roller {
	t.Helper()
	r, err := NewRollerURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
