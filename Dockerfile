# ---- build ----
FROM golang:1.23-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/shortener .

# ---- run ----
FROM alpine:3.20

# Needed for TLS to Neon and Upstash.
RUN apk add --no-cache ca-certificates

COPY --from=build /out/shortener /usr/local/bin/shortener

# Overridden by the host; the app reads $PORT.
ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/shortener"]