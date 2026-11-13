# gnoshard

A Cosmos SDK chain that runs the Gno virtual machine and leases its validator
set from a VAAS provider. It is the vaas repository's reference consumer
(`app/consumer`) with [ignite/gnovm](https://github.com/ignite/gnovm)'s
`x/gnovm` added: auth, bank, vesting, params, consensus, upgrade, genutil, ibc
core, transfer, `vaasconsumer` and `gnovm`. There is no staking, slashing or
gov module: the validator set arrives over VSC packets, and the chain's
authority is its owner on the provider, acting through `MsgOwnerExec`.

The name, the bech32 prefix (`shard`) and the fee denom (`ushard`) are
placeholders, each defined once in `app/chain.go`.

## Build and test

```bash
make build        # build/gnoshardd
make test
make docker-build
```

Go 1.26.11. The module pins atomone-sdk the way vaas does, `github.com/allinbits/vaas`
replaced by the `pre-mvp` branch of giunatale/vaas until a release,
`github.com/ignite/gnovm` replaced by the `gnoroot` branch of giunatale/gnovm
until its change (the standard library located through `GNOROOT`) is
upstream, and the gno fork ignite/gnovm's README names.

## The Gno standard library

`x/gnovm` loads the Gno standard library at InitGenesis from
`GNOROOT/gnovm/stdlibs`. With `GNOROOT` unset it asks the Go toolchain for the
gno module's directory, which works in a checkout and nowhere else; a node
image sets `GNOROOT` and carries a copy of the library under it, as the
Dockerfile does. The localnet targets set it from the module cache.

## Localnet, as a consumer of a provider on this host

```bash
make localnet-init                     # ~/.gnoshard-localnet, chain id gnoshard-localnet
# create the consumer on the provider with chain id gnoshard-localnet
make localnet-genesis CONSUMER_ID=<id> providerd="<provider binary> --home <provider home>"
make localnet-run provider_home=<provider home>
```

From there the flow is any consumer's (see the consumer launch runbook in the
vaas repository): a relayer creates the IBC v2 clients, the owner declares
them on both chains, the fee pool is funded, and the validator set arrives.

## Using the VM

```bash
gnoshardd tx gnovm add-package ./counter --from user --gas 5000000 --fees 50000ushard
gnoshardd tx gnovm call gno.land/r/demo/counter Increment --from user --gas 5000000 --fees 50000ushard
gnoshardd query gnovm render gno.land/r/demo/counter ''
gnoshardd query gnovm eval gno.land/r/demo/counter 'Render("")'
```

Gno messages cannot share a transaction with other messages, and the VM's
storage deposit and price are quoted in the chain's denom (set at genesis;
`localnet-init` does it).
