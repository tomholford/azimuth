package azimuth

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestIsL1Dominion(t *testing.T) {
	if !IsL1Dominion(DominionL1) {
		t.Fatal("l1")
	}
	if IsL1Dominion(DominionL2) || IsL1Dominion(DominionSpawn) || IsL1Dominion("") {
		t.Fatal("non-l1")
	}
}

func TestDominionFromDeed(t *testing.T) {
	owner := common.HexToAddress("0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	proxy := common.HexToAddress("0xF7306c5db0C1880FB2ed9c3972ad3e1A94999196")
	deposit := DepositAddr()
	tests := []struct {
		owner, spawn common.Address
		want         string
	}{
		{owner: owner, spawn: proxy, want: DominionL1},
		{owner: owner, spawn: common.Address{}, want: DominionL1},
		{owner: common.Address{}, spawn: common.Address{}, want: DominionL1},
		{owner: deposit, spawn: common.Address{}, want: DominionL2},
		{owner: deposit, spawn: deposit, want: DominionL2},
		{owner: owner, spawn: deposit, want: DominionSpawn},
	}
	for _, tc := range tests {
		got := DominionFromDeed(tc.owner, tc.spawn)
		if got != tc.want {
			t.Errorf("owner=%s spawn=%s: %s want %s", tc.owner.Hex(), tc.spawn.Hex(), got, tc.want)
		}
	}
}

func TestKeysWritePath(t *testing.T) {
	testWritePath(t, KeysWritePath, map[string]string{
		DominionL1:    KeysPathEcliptic,
		DominionSpawn: KeysPathEcliptic,
		DominionL2:    KeysPathRoller,
	})
}

func TestSpawnWritePath(t *testing.T) {
	testWritePath(t, SpawnWritePath, map[string]string{
		DominionL1:    KeysPathEcliptic,
		DominionSpawn: KeysPathRoller,
		DominionL2:    KeysPathRoller,
	})
}

func testWritePath(t *testing.T, fn func(string) (string, error), want map[string]string) {
	t.Helper()
	tests := []struct {
		dom     string
		want    string
		wantErr string
	}{
		{dom: DominionL1, want: want[DominionL1]},
		{dom: DominionSpawn, want: want[DominionSpawn]},
		{dom: DominionL2, want: want[DominionL2]},
		{dom: "", wantErr: "unknown dominion"},
		{dom: "other", wantErr: "unsupported dominion other"},
	}
	for _, tc := range tests {
		got, err := fn(tc.dom)
		if tc.wantErr != "" {
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("%q: err=%v want %q", tc.dom, err, tc.wantErr)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Fatalf("%q: got %q err=%v want %q", tc.dom, got, err, tc.want)
		}
	}
}
