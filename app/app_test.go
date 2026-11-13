package app

import (
	"encoding/json"
	"testing"

	gnovmtypes "github.com/ignite/gnovm/x/gnovm/types"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	consumertypes "github.com/allinbits/vaas/x/vaas/consumer/types"
	vaastypes "github.com/allinbits/vaas/x/vaas/types"
)

const testChainID = "gnoshard-test"

// TestInitChainLoadsTheGnoVM: a genesis carrying the one validator a consumer
// needs initializes every module, the Gno VM included, which loads its
// standard library at that point; the chain then finalizes and commits a
// block.
func TestInitChainLoadsTheGnoVM(t *testing.T) {
	app := New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, sims.EmptyAppOptions{}, baseapp.SetChainID(testChainID))

	genesisState := NewDefaultGenesisState(app.AppCodec())
	valPubKey, err := cryptocodec.ToCmtProtoPublicKey(ed25519.GenPrivKey().PubKey())
	require.NoError(t, err)
	consumerGenesis := consumertypes.NewRestartGenesisState(
		"",
		[]abci.ValidatorUpdate{{PubKey: valPubKey, Power: 100}},
		vaastypes.DefaultConsumerParams(),
	)
	genesisState[consumertypes.ModuleName] = app.AppCodec().MustMarshalJSON(consumerGenesis)
	stateBytes, err := json.Marshal(genesisState)
	require.NoError(t, err)

	_, err = app.InitChain(&abci.RequestInitChain{
		ChainId:         testChainID,
		ConsensusParams: sims.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
		InitialHeight:   1,
	})
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	ctx := app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight(), ChainID: testChainID})
	params, err := app.GnovmKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, gnovmtypes.DefaultParams(), params)
}

// TestTheGnoVMIsWiredLikeTheOtherModules: the VM runs in every module order,
// mints through its own module account, and takes its parameter changes from
// the owner's authority.
func TestTheGnoVMIsWiredLikeTheOtherModules(t *testing.T) {
	app := New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, sims.EmptyAppOptions{})

	require.Contains(t, app.MM.Modules, gnovmtypes.ModuleName)
	require.Contains(t, app.MM.OrderBeginBlockers, gnovmtypes.ModuleName)
	require.Contains(t, app.MM.OrderEndBlockers, gnovmtypes.ModuleName)
	require.Contains(t, app.MM.OrderInitGenesis, gnovmtypes.ModuleName)
	require.Equal(t, []string{authtypes.Minter}, GetMaccPerms()[gnovmtypes.ModuleName])
	require.Equal(t, []byte(consumertypes.OwnerExecAuthority()), app.GnovmKeeper.GetAuthority())
	require.NotNil(t, app.GetKey(gnovmtypes.ModuleName))
	require.NotNil(t, app.GetMemKey(gnovmMemStoreKey))
}
