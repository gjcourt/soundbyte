FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Server: Pure Go — cross-compile for target platform
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

# Final Stage
FROM alpine:latest

COPY --from=builder /server /usr/local/bin/soundbyte-server

# Default to server
ENTRYPOINT ["soundbyte-server"]
