package azimuth

import "testing"

func TestIsL1Dominion(t *testing.T) {
	if !IsL1Dominion(DominionL1) {
		t.Fatal("l1")
	}
	if IsL1Dominion(DominionL2) || IsL1Dominion(DominionSpawn) || IsL1Dominion("") {
		t.Fatal("non-l1")
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
