# ---- build stage ----
FROM golang:1.25-alpine AS builder
WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

# Build server and client binaries. The fingerprint rule set is embedded
# (go:embed), so no runtime config files are required.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/client ./cmd/client

# ---- runtime stage ----
FROM alpine:3.20
RUN adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=builder /out/server /out/client /app/
# Copy sample data so the client can self-validate inside the container.
COPY testdata/ /data/
USER appuser
EXPOSE 8080
CMD ["/app/server", "-addr", ":8080"]
