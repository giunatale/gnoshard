package params_test

import (
	"testing"

	gnocrypto "github.com/gnolang/gno/tm2/pkg/crypto"
	_ "github.com/ignite/gnovm/x/gnovm/keeper"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/giunatale/gnoshard/app/params"
)

// TestTheGnoVMTakesTheChainPrefix: the VM copies the prefix this package set,
// not cosmos-sdk's default, which depends on this package initializing
// before x/gnovm's keeper package.
func TestTheGnoVMTakesTheChainPrefix(t *testing.T) {
	require.Equal(t, params.DefaultBech32Prefix, sdk.GetConfig().GetBech32AccountAddrPrefix())
	require.Equal(t, params.DefaultBech32Prefix, gnocrypto.Bech32AddrPrefix())
	require.Equal(t, params.DefaultBech32Prefix+"pub", gnocrypto.Bech32PubKeyPrefix())
}
