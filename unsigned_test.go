package azimuth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestSubmitL2Errors(t *testing.T) {
	r := NewRoller()
	sig := "0x00"
	addr := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	ctx := context.Background()

	if _, err := (*Unsigned)(nil).SubmitL2(ctx, r, sig, addr); err == nil {
		t.Fatal("nil unsigned")
	}
	u := &Unsigned{Layer: LayerL2, Tx: L2TxSpawn, Payload: SpawnData{}}
	if _, err := u.SubmitL2(ctx, nil, sig, addr); err == nil {
		t.Fatal("nil roller")
	}
	u = &Unsigned{Layer: LayerL1, Tx: L2TxSpawn, Payload: SpawnData{}}
	if _, err := u.SubmitL2(ctx, r, sig, addr); err == nil || !strings.Contains(err.Error(), "layer") {
		t.Fatalf("l1: %v", err)
	}
	u = &Unsigned{Layer: LayerL2, Tx: "nope", Payload: SpawnData{}}
	if _, err := u.SubmitL2(ctx, r, sig, addr); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("unknown: %v", err)
	}
	u = &Unsigned{Layer: LayerL2, Tx: L2TxSpawn, Payload: AddressData{}}
	if _, err := u.SubmitL2(ctx, r, sig, addr); err == nil || !strings.Contains(err.Error(), "payload") {
		t.Fatalf("payload: %v", err)
	}
}

func TestSubmitL2Dispatch(t *testing.T) {
	from := L2From{Ship: "~zod", Proxy: ProxyOwn}
	addr := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	sig := "0x" + strings.Repeat("cd", 65)
	proxy := AddressData{Address: "0xF7306c5db0C1880FB2ed9c3972ad3e1A94999196"}
	spawn, err := NewSpawnData(69, common.HexToAddress(proxy.Address))
	if err != nil {
		t.Fatal(err)
	}
	ship, err := NewShipData(256)
	if err != nil {
		t.Fatal(err)
	}

	tests := []*Unsigned{
		{Layer: LayerL2, Tx: L2TxSetManagementProxy, From: from, Payload: proxy},
		{Layer: LayerL2, Tx: L2TxSpawn, From: from, Payload: spawn},
		{Layer: LayerL2, Tx: L2TxEscape, From: from, Payload: ship},
	}
	for _, u := range tests {
		t.Run(u.Tx, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req struct {
					Method string `json:"method"`
					Params struct {
						Sig     string `json:"sig"`
						Address string `json:"address"`
					} `json:"params"`
				}
				if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
					t.Errorf("decode: %v", decErr)
					return
				}
				if req.Method != u.Tx {
					t.Errorf("method %s want %s", req.Method, u.Tx)
				}
				if req.Params.Sig != sig || req.Params.Address != addr {
					t.Errorf("params %+v", req.Params)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1337","result":"0xbeef"}`))
			}))
			defer srv.Close()

			got, err := u.SubmitL2(context.Background(), mustRoller(t, srv.URL), sig, addr)
			if err != nil {
				t.Fatal(err)
			}
			if got != "0xbeef" {
				t.Fatalf("hash %q", got)
			}
		})
	}
}
