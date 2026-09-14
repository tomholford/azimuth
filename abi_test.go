package azimuth

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestAzimuthABI(t *testing.T) {
	parsed, err := AzimuthABI()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"getKeys", "getOwnedPoints", "getOwner", "canManage", "getManagerFor", "owner",
		"getManagementProxy", "getSpawnProxy", "getTransferProxy", "getVotingProxy",
		"points", "rights", "getSpawned", "getSponsoring", "getEscapeRequests",
		"getSpawningFor", "getTransferringFor", "getVotingFor",
	} {
		if _, ok := parsed.Methods[name]; !ok {
			t.Errorf("missing %s", name)
		}
	}
	if AzimuthAddr() == (common.Address{}) {
		t.Fatal("zero azimuth address")
	}
}

func TestEclipticABI(t *testing.T) {
	parsed, err := EclipticABI()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"configureKeys", "setManagementProxy", "setSpawnProxy", "setTransferProxy",
		"setVotingProxy", "spawn", "transferPoint", "escape", "cancelEscape",
		"adopt", "reject", "detach", "owner", "depositAddress",
	} {
		if _, ok := parsed.Methods[name]; !ok {
			t.Errorf("missing %s", name)
		}
	}
	want := common.HexToAddress(EclipticAddress)
	if EclipticAddr() != want || want == (common.Address{}) {
		t.Fatalf("address %s", EclipticAddr().Hex())
	}
}

func TestClaimsABI(t *testing.T) {
	parsed, err := ClaimsABI()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"addClaim", "removeClaim", "clearClaims", "findClaim", "claims",
	} {
		if _, ok := parsed.Methods[name]; !ok {
			t.Errorf("missing %s", name)
		}
	}
	want := common.HexToAddress(ClaimsAddress)
	if ClaimsAddr() != want || want == (common.Address{}) {
		t.Fatalf("address %s", ClaimsAddr().Hex())
	}
}

func TestDepositAddr(t *testing.T) {
	want := common.HexToAddress(DepositAddress)
	if DepositAddr() != want || want == (common.Address{}) {
		t.Fatalf("address %s", DepositAddr().Hex())
	}
	if DepositAddr() == AzimuthAddr() || DepositAddr() == EclipticAddr() {
		t.Fatal("deposit address collides with a core contract")
	}
}

func TestNewContracts(t *testing.T) {
	az, err := NewAzimuthContract()
	if err != nil {
		t.Fatal(err)
	}
	if az.Address != AzimuthAddr() {
		t.Fatalf("azimuth %s", az.Address.Hex())
	}
	ec, err := NewEclipticContract()
	if err != nil {
		t.Fatal(err)
	}
	if ec.Address != EclipticAddr() {
		t.Fatalf("ecliptic %s", ec.Address.Hex())
	}
	cl, err := NewClaimsContract()
	if err != nil {
		t.Fatal(err)
	}
	if cl.Address != ClaimsAddr() {
		t.Fatalf("claims %s", cl.Address.Hex())
	}
}
