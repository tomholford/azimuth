package azimuth

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func loadJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(err)
	}
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	s = strings.TrimPrefix(s, "0x")
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func mockEthRPC(t *testing.T, methods map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reply := func(raw json.RawMessage) json.RawMessage {
			var req struct {
				ID     json.RawMessage `json:"id"`
				Method string          `json:"method"`
			}
			if err := json.Unmarshal(raw, &req); err != nil {
				return []byte(`{"jsonrpc":"2.0","error":{"code":-32700,"message":"parse"}}`)
			}
			v, ok := methods[req.Method]
			if !ok {
				if req.Method == "eth_chainId" || req.Method == "net_version" {
					v = "0x1"
				} else {
					return mustJSON(map[string]any{
						"jsonrpc": "2.0",
						"id":      req.ID,
						"error":   map[string]any{"code": -32601, "message": "method not found: " + req.Method},
					})
				}
			}
			return mustJSON(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": v})
		}
		w.Header().Set("Content-Type", "application/json")
		if len(body) > 0 && body[0] == '[' {
			var batch []json.RawMessage
			if err := json.Unmarshal(body, &batch); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			out := make([]json.RawMessage, len(batch))
			for i, item := range batch {
				out[i] = reply(item)
			}
			_ = json.NewEncoder(w).Encode(out)
			return
		}
		_, _ = w.Write(reply(body))
	}))
}

func mockJSONRPC(t *testing.T, handle func(method string, params json.RawMessage) (any, error)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		result, err := handle(req.Method, req.Params)
		if err != nil {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"error":   map[string]any{"code": -32000, "message": err.Error()},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  result,
		})
	}))
}

func ethCallData(params json.RawMessage) string {
	var p []json.RawMessage
	if err := json.Unmarshal(params, &p); err != nil || len(p) == 0 {
		return ""
	}
	var args struct {
		Data  string `json:"data"`
		Input string `json:"input"`
	}
	if err := json.Unmarshal(p[0], &args); err != nil {
		return ""
	}
	if args.Input != "" {
		return args.Input
	}
	return args.Data
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
