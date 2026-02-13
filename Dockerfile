FROM golang:1.23-alpine AS builder

# Install dependencies for Client (ALSA)
RUN apk add --no-cache gcc musl-dev alsa-lib-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Server: Pure Go (Pion Opus is pure)
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

# Client: Requires CGO for ALSA (Linux audio)
RUN CGO_ENABLED=1 go build -o /client ./cmd/client

# Final Stage
FROM alpine:latest
RUN apk add --no-cache alsa-lib

COPY --from=builder /server /usr/local/bin/soundbyte-server
COPY --from=builder /client /usr/local/bin/soundbyte-client

# Default to server
ENTRYPOINT ["soundbyte-server"]
