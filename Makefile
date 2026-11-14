BUILDDIR ?= $(CURDIR)/build
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null)
COMMIT := $(shell git log -1 --format='%H' 2>/dev/null)

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=gnoshard \
	-X github.com/cosmos/cosmos-sdk/version.AppName=gnoshardd \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

###############################################################################
###                                  Build                                  ###
###############################################################################

build: go.sum
	mkdir -p $(BUILDDIR)
	go build -mod=readonly -ldflags "$(ldflags)" -o $(BUILDDIR)/ ./cmd/gnoshardd

install: go.sum
	go install -mod=readonly -ldflags "$(ldflags)" ./cmd/gnoshardd

test:
	go test -timeout 20m ./...

vet:
	go vet ./...

docker-build:
	docker build -t gnoshard:latest .

.PHONY: build install test vet docker-build

###############################################################################
###                                 Localnet                                ###
###############################################################################
#
# One gnoshard node as a VAAS consumer of a provider already running on this
# host. The flow mirrors the vaas repository's consumer targets:
#
# The node's home is build/localnet (override localnet_home); nothing is
# written under $HOME.
#
#   make localnet-init
#   <create the consumer on the provider with chain id $(chain_id)>
#   make localnet-genesis CONSUMER_ID=<id> providerd="<provider binary> --home <provider home>"
#   make localnet-run provider_home=<provider home>
#
# The node then waits for the relayer's clients and the owner's declarations,
# like any consumer (see the vaas consumer launch runbook).

localnet_home ?= $(BUILDDIR)/localnet
chain_id ?= gnoshard-localnet
denom ?= ushard
provider_home ?= $(HOME)/.provider-localnet
providerd ?= provider --home $(provider_home)
CONSUMER_ID ?= 1
localnetd = $(BUILDDIR)/gnoshardd --home $(localnet_home)

# The Gno root directory, whose gnovm/stdlibs the node loads at InitGenesis:
# the gno module's copy in the module cache, named explicitly so the node
# does not depend on `go list` resolving it from its working directory.
GNOROOT ?= $(shell go list -m -f '{{.Dir}}' github.com/gnolang/gno)
localnet_env = GNOROOT=$(GNOROOT)

# Ports offset from the defaults so the node coexists with a provider and the
# vaas reference consumer on one host.
rpc_port ?= 26677
p2p_port ?= 26676
abci_port ?= 26678
grpc_port ?= 9093
api_port ?= 1319
pprof_port ?= 6063

localnet-init: build
	rm -rf $(localnet_home)
	$(localnetd) init localnet --default-denom $(denom) --chain-id $(chain_id)
	$(localnetd) config set client chain-id $(chain_id)
	$(localnetd) config set client keyring-backend test
	$(localnetd) keys add user
	./scripts/add-relayer-key.sh $(BUILDDIR)/gnoshardd $(localnet_home)
	./scripts/add-owner-key.sh $(BUILDDIR)/gnoshardd $(localnet_home)
	$(localnetd) genesis add-genesis-account user 1000000000$(denom)
	$(localnetd) genesis add-genesis-account relayer 100000000$(denom)
	$(localnetd) genesis add-genesis-account owner 100000000$(denom)
	# x/gnovm's default deposit and storage price are quoted in cosmos-sdk's
	# default bond denom; the chain pays them in its own.
	jq '.app_state.gnovm.params.default_deposit |= sub("[a-z]+$$"; "$(denom)") | .app_state.gnovm.params.storage_price |= sub("[a-z]+$$"; "$(denom)")' \
		$(localnet_home)/config/genesis.json > $(localnet_home)/config/genesis.json.tmp
	mv $(localnet_home)/config/genesis.json.tmp $(localnet_home)/config/genesis.json
	# genesis validate waits for localnet-genesis: a consumer genesis without
	# the provider's payload has no validator set and is refused.
	sed -i.bak 's#^minimum-gas-prices = .*#minimum-gas-prices = "0.01$(denom)"#g' $(localnet_home)/config/app.toml
	sed -i.bak 's#tcp://127.0.0.1:26657#tcp://0.0.0.0:$(rpc_port)#g' $(localnet_home)/config/config.toml
	sed -i.bak 's#tcp://0.0.0.0:26656#tcp://0.0.0.0:$(p2p_port)#g' $(localnet_home)/config/config.toml
	sed -i.bak 's#tcp://127.0.0.1:26658#tcp://127.0.0.1:$(abci_port)#g' $(localnet_home)/config/config.toml
	sed -i.bak 's#localhost:6060#localhost:$(pprof_port)#g' $(localnet_home)/config/config.toml
	sed -i.bak 's#tcp://localhost:26657#tcp://localhost:$(rpc_port)#g' $(localnet_home)/config/client.toml
	sed -i.bak 's#localhost:9090#localhost:$(grpc_port)#g' $(localnet_home)/config/app.toml
	sed -i.bak 's#localhost:1317#localhost:$(api_port)#g' $(localnet_home)/config/app.toml
	@echo "Node initialized. Create the consumer on the provider with chain id $(chain_id), then run 'make localnet-genesis CONSUMER_ID=<id>'."

localnet-genesis:
	$(providerd) query vaasprovider consumer-genesis $(CONSUMER_ID) -o json > $(localnet_home)/consumer_genesis.json
	jq --slurpfile cg $(localnet_home)/consumer_genesis.json '.app_state.vaasconsumer = $$cg[0]' \
		$(localnet_home)/config/genesis.json > $(localnet_home)/config/genesis.json.tmp
	mv $(localnet_home)/config/genesis.json.tmp $(localnet_home)/config/genesis.json
	$(localnetd) genesis validate
	@echo "Consumer genesis patched. Run 'make localnet-run'."

# A single-validator localnet signs with the provider's validator key, since
# the consumer's validator set is the provider's.
localnet-copy-val-key:
	cp $(provider_home)/config/priv_validator_key.json $(localnet_home)/config/priv_validator_key.json
	cp $(provider_home)/config/node_key.json $(localnet_home)/config/node_key.json

localnet-run: localnet-copy-val-key
	$(localnet_env) $(localnetd) start

localnet-clean:
	rm -rf $(localnet_home)

.PHONY: localnet-init localnet-genesis localnet-copy-val-key localnet-run localnet-clean
