FROM --platform=$BUILDPLATFORM golang:1.22-bookworm AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -trimpath -ldflags="-s -w -X main.version=${SOURCEVERSION:-dev}" \
    -o /out/gateway ./cmd/gateway

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/gateway /app/gateway
COPY config.example.yaml /app/config.example.yaml
EXPOSE 9091
USER nonroot:nonroot
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD ["/app/gateway", "doctor", "--config=/app/config.example.yaml"]
ENTRYPOINT ["/app/gateway"]
