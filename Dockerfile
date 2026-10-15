# The Go version matches the toolchain line in go.mod, so the image builds with the Go that CI tests with.
FROM golang:1.27.2-trixie AS build

WORKDIR /src
COPY . .

ARG VERSION
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.local/share/pnpm/store \
    --mount=type=cache,target=/src/.vite-plus \
    make build VERSION="${VERSION:?pass --build-arg VERSION, as make image does}"

# The nonroot tag runs as uid 65532 and ships CA certificates and tzdata.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab

COPY --from=build /src/bin/web /src/bin/migrate /app/

# Lets the platform override the command alone, without having to reset an entrypoint.
ENTRYPOINT []
CMD ["/app/web"]
