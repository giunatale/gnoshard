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
	-o /out/ ./cmd/gnoshardd

# The Gno standard library under GNOROOT: gnovm/stdlibs, which the node loads
# at InitGenesis, and gnovm/tests/stdlibs, which the VM reads when it
# type-checks the library's test files.
RUN gno="$(go list -m -f '{{.Dir}}' github.com/gnolang/gno)" && mkdir -p /out/gnoroot/gnovm/tests && \
    cp -r "$gno/gnovm/stdlibs" /out/gnoroot/gnovm/stdlibs && \
    cp -r "$gno/gnovm/tests/stdlibs" /out/gnoroot/gnovm/tests/stdlibs

FROM alpine:${IMG_TAG}

RUN apk add --no-cache bash curl jq
RUN adduser -D nonroot -h /home/nonroot

COPY --from=builder /out/gnoshardd /usr/local/bin/
COPY --from=builder /out/gnoroot /usr/local/share/gno
ENV GNOROOT=/usr/local/share/gno

EXPOSE 26656 26657 1317 9090

USER nonroot
ENTRYPOINT ["gnoshardd"]
