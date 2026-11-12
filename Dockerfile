# syntax=docker/dockerfile:1.28

FROM --platform=$BUILDPLATFORM golang:1.27.4 AS builder
WORKDIR /src

ENV GOPROXY=https://proxy.golang.org,direct

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w" -o /out/server ./cmd/server

FROM alpine:3.24

ARG APP_UID=10001
ARG APP_GID=10001

RUN addgroup -S -g "${APP_GID}" app \
    && adduser -S -D -H -u "${APP_UID}" -G app app

COPY --from=builder --chown=${APP_UID}:${APP_GID} --chmod=0555 /out/server /usr/local/bin/server

USER ${APP_UID}:${APP_GID}

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/server"]
