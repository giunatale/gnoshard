ARG GO_VERSION=1.26.11
ARG IMG_TAG=3.22

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache git ca-certificates build-base linux-headers

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
RUN CGO_ENABLED=0 go build -mod=readonly \
	-ldflags "-X github.com/cosmos/cosmos-sdk/version.Name=gnoshard -X github.com/cosmos/cosmos-sdk/version.AppName=gnoshardd -X github.com/cosmos/cosmos-sdk/version.Version=${VERSION} -X github.com/cosmos/cosmos-sdk/version.Commit=${COMMIT}" \
	-o /out/ ./cmd/gnoshardd ./cmd/gnostdlibs-driver

# The Gno standard library, which the node loads at InitGenesis. x/gnovm asks
# the Go toolchain where the gno module keeps it; the image carries a copy and
# gnostdlibs-driver answers in the toolchain's place.
RUN cp -r "$(go list -m -f '{{.Dir}}' github.com/gnolang/gno)/gnovm/stdlibs" /out/gno-stdlibs

FROM alpine:${IMG_TAG}

RUN apk add --no-cache bash curl jq
RUN adduser -D nonroot -h /home/nonroot

COPY --from=builder /out/gnoshardd /out/gnostdlibs-driver /usr/local/bin/
COPY --from=builder /out/gno-stdlibs /usr/local/share/gno/stdlibs
ENV GOPACKAGESDRIVER=/usr/local/bin/gnostdlibs-driver \
	GNO_STDLIBS_DIR=/usr/local/share/gno/stdlibs

EXPOSE 26656 26657 1317 9090

USER nonroot
ENTRYPOINT ["gnoshardd"]
