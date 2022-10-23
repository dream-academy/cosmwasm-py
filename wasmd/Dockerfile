# docker build . -t cosmwasm/wasmd:latest
# docker run --rm -it cosmwasm/wasmd:latest /bin/sh
FROM golang:1.17-alpine3.15 AS go-builder
ARG arch=x86_64

# this comes from standard alpine nightly file
#  https://github.com/rust-lang/docker-rust-nightly/blob/master/alpine3.12/Dockerfile
# with some changes to support our toolchain, etc
RUN set -eux; apk add --no-cache ca-certificates build-base;

RUN apk add git

WORKDIR /code
COPY . /code

# See https://github.com/CosmWasm/wasmvm/releases
COPY ./libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.a
COPY ./libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a

RUN LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make libwasmd