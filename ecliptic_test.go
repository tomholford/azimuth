package azimuth

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

func TestPackConfigureKeys(t *testing.T) {
	crypt := bytes.Repeat([]byte{0xaa}, 32)
	auth := bytes.Repeat([]byte{0xbb}, 32)
	data, err := PackConfigureKeys(69, crypt, auth, CryptoSuiteVersion, true)
	if err != nil {
		t.Fatal(err)
	}
	sel := crypto.Keccak256([]byte("configureKeys(uint32,bytes32,bytes32,uint32,bool)"))[:4]
	if !bytes.Equal(data[:4], sel) {
		t.Fatalf("selector %x want %x", data[:4], sel)
	}
	parsed, err := EclipticABI()
	if err != nil {
		t.Fatal(err)
	}
	unpacked, err := parsed.Methods["configureKeys"].Inputs.Unpack(data[4:])
	if err != nil {
		t.Fatal(err)
	}
	if unpacked[0].(uint32) != 69 {
		t.Fatalf("point %v", unpacked[0])
	}
	gotCrypt := unpacked[1].([32]byte)
	gotAuth := unpacked[2].([32]byte)
	if !bytes.Equal(gotCrypt[:], crypt) || !bytes.Equal(gotAuth[:], auth) {
		t.Fatalf("keys crypt=%x auth=%x", gotCrypt, gotAuth)
	}
	if unpacked[3].(uint32) != 1 {
		t.Fatalf("suite %v", unpacked[3])
	}
	if unpacked[4].(bool) != true {
		t.Fatal("discontinuous")
	}
}

func TestPackConfigureKeysBadLen(t *testing.T) {
	_, err := PackConfigureKeys(1, []byte{1}, bytes.Repeat([]byte{2}, 32), 1, false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPackConfigureKeysHexRoundTrip(t *testing.T) {
	crypt, err := hex.DecodeString("48d5ddd4061257ba166fa3f9bbdb74f1a4e81c089384fa77f790709f0dfbc766")
	if err != nil {
		t.Fatal(err)
	}
	auth, err := hex.DecodeString("2b42105eb8900aea081451ad7980e4adaa762d68505f79d3aab8f8bafe04eec1")
	if err != nil {
		t.Fatal(err)
	}
	data, err := PackConfigureKeys(69, crypt, auth, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 4+32*5 {
		t.Fatalf("len %d", len(data))
	}
}

func TestPackConfigureKeysGolden(t *testing.T) {
	var fx struct {
		Input string `json:"input"`
		Args  struct {
			AuthenticationKey  string `json:"authenticationKey"`
			CryptoSuiteVersion uint32 `json:"cryptoSuiteVersion"`
			Discontinuous      bool   `json:"discontinuous"`
			EncryptionKey      string `json:"encryptionKey"`
			Point              uint32 `json:"point"`
		} `json:"args"`
	}
	loadJSON(t, "testdata/l1/configureKeys-622dd42b.json", &fx)
	enc := mustHex(t, fx.Args.EncryptionKey)
	auth := mustHex(t, fx.Args.AuthenticationKey)
	got, err := PackConfigureKeys(fx.Args.Point, enc, auth, fx.Args.CryptoSuiteVersion, fx.Args.Discontinuous)
	if err != nil {
		t.Fatal(err)
	}
	want := mustHex(t, fx.Input)
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x want %x", got, want)
	}
}
