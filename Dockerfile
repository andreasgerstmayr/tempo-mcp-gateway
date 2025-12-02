FROM golang:1.24 AS builder
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build .

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/tempo-mcp-gateway .
USER 65532:65532

ENTRYPOINT ["/tempo-mcp-gateway"]
