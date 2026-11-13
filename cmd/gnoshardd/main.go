package main

import (
	"fmt"
	"os"

	gnocrypto "github.com/gnolang/gno/tm2/pkg/crypto"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdk "github.com/cosmos/cosmos-sdk/types"

	app "github.com/giunatale/gnoshard/app"
	// app/params sets the chain's address prefix as it initializes; see its
	// package comment.
	_ "github.com/giunatale/gnoshard/app/params"
	"github.com/giunatale/gnoshard/cmd/gnoshardd/cmd"
)

func main() {
	if sdkPrefix, vmPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix(), gnocrypto.Bech32AddrPrefix(); sdkPrefix != vmPrefix {
		fmt.Fprintf(os.Stderr, "the Gno VM took the address prefix %q before the chain set %q\n", vmPrefix, sdkPrefix)
		os.Exit(1)
	}

	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
