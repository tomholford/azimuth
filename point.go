package azimuth

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/deelawn/urbit-gob/co"
	"github.com/ethereum/go-ethereum/common"
)

//
// Types
//

// Proxy is an Ethereum address plus the roller nonce for that role.
type Proxy struct {
	Address common.Address `json:"address"`
	Nonce   int            `json:"nonce"`
}

// Point is live Azimuth state for one ship. No cache timestamp — callers
// treat it as a snapshot of the last read.
type Point struct {
	Index    uint32 `json:"point"`
	Name     string `json:"name"`
	Dominion string `json:"dominion"`

	Owner           Proxy `json:"owner"`
	ManagementProxy Proxy `json:"managementProxy"`
	SpawnProxy      Proxy `json:"spawnProxy"`
	TransferProxy   Proxy `json:"transferProxy"`
	VotingProxy     Proxy `json:"votingProxy"`

	CryptKey []byte `json:"crypt_key"`
	EdKey    []byte `json:"ed_key"`
	Suite    uint32 `json:"crypto_suite_version"`
	Revision uint32 `json:"key_revision_number"`
	Rift     uint32 `json:"rift"`

	HasSponsor        bool   `json:"has_sponsor"`
	Sponsor           uint32 `json:"sponsor"`
	Active            bool   `json:"active"`
	EscapeRequested   bool   `json:"escape_requested"`
	EscapeRequestedTo uint32 `json:"escape_requested_to"`
}

//
// Public
//

// ParsePoint parses a @p (with or without leading ~) to an Azimuth point number.
// Moons and comets are rejected — they are not on Azimuth.
func ParsePoint(name string) (uint32, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("empty @p")
	}
	if !strings.HasPrefix(name, "~") {
		name = "~" + name
	}
	if !co.IsValidPatp(name) {
		return 0, fmt.Errorf("invalid @p %q", name)
	}
	dec, err := co.Patp2Dec(name)
	if err != nil {
		return 0, fmt.Errorf("invalid @p %q: %w", name, err)
	}
	n, err := strconv.ParseUint(dec, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s: not an Azimuth point: %w", name, err)
	}
	return uint32(n), nil
}

// FormatPoint renders an Azimuth point number as @p (with ~).
func FormatPoint(n uint32) (string, error) {
	s, err := co.Patp(strconv.FormatUint(uint64(n), 10))
	if err != nil {
		return "", fmt.Errorf("format point %d: %w", n, err)
	}
	return s, nil
}

// Clan is galaxy / star / planet for an Azimuth point.
func Clan(n uint32) (string, error) {
	name, err := FormatPoint(n)
	if err != nil {
		return "", err
	}
	return co.Clan(name)
}

// Prefix is the ship's parent on the network (galaxy→self, star→galaxy, planet→star).
func Prefix(n uint32) (uint32, error) {
	name, err := FormatPoint(n)
	if err != nil {
		return 0, err
	}
	parent, err := co.Sein(name)
	if err != nil {
		return 0, fmt.Errorf("prefix of %s: %w", name, err)
	}
	return ParsePoint(parent)
}
