// Package params sets the chain's bech32 address prefix.
//
// It does so from its init function, before main runs, because x/gnovm's
// keeper package copies cosmos-sdk's prefixes into the Gno VM from its own
// init function and the VM refuses a second setting. Go initializes the two
// packages in import-path order once their imports are ready, and this one
// sorts first; cmd/gnoshardd checks that the two prefixes agree.
package params

import (
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/allinbits/vaas/app/bech32prefix"
)

const (
	// DefaultBech32Prefix is the address prefix when Bech32PrefixEnvVar is
	// unset.
	DefaultBech32Prefix = "gnoshard"

	// Bech32PrefixEnvVar names the environment variable that overrides the
	// address prefix, for test networks.
	Bech32PrefixEnvVar = "GNOSHARD_BECH32_PREFIX"
)

func init() {
	prefix := DefaultBech32Prefix
	if v, ok := os.LookupEnv(Bech32PrefixEnvVar); ok {
		if err := bech32prefix.Validate(v); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", Bech32PrefixEnvVar, err)
			os.Exit(1)
		}
		prefix = v
	}
	SetAddressPrefixes(prefix)
}

// SetAddressPrefixes builds the Config with Bech32 addressPrefix and publKeyPrefix for accounts, validators, and consensus nodes and verifies that addreeses have correct format.
func SetAddressPrefixes(bech32Prefix string) {
	cfg := sdk.GetConfig()

	// bech32PrefixAccAddr defines the Bech32 prefix of an account's address
	bech32PrefixAccAddr := bech32Prefix
	// bech32PrefixAccPub defines the Bech32 prefix of an account's public key
	bech32PrefixAccPub := bech32Prefix + sdk.PrefixPublic
	// bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	bech32PrefixValAddr := bech32Prefix + sdk.PrefixValidator + sdk.PrefixOperator
	// bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	bech32PrefixValPub := bech32Prefix + sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic
	// bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	bech32PrefixConsAddr := bech32Prefix + sdk.PrefixValidator + sdk.PrefixConsensus
	// bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	bech32PrefixConsPub := bech32Prefix + sdk.PrefixValidator + sdk.PrefixConsensus + sdk.PrefixPublic

	cfg.SetBech32PrefixForAccount(bech32PrefixAccAddr, bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(bech32PrefixValAddr, bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(bech32PrefixConsAddr, bech32PrefixConsPub)

	cfg.Seal()
}
