ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app ./cmd/pokerhans


FROM debian:bookworm

# Installiere notwendige Pakete
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Kopiere Binary
COPY --from=builder /run-app /usr/local/bin/

# Kopiere Templates und statische Dateien
# Hinweis: Die CSS-Dateien werden bereits im GitHub Actions Workflow gebaut
COPY --from=builder /usr/src/app/web /app/web
COPY --from=builder /usr/src/app/migrations /app/migrations

# Arbeitsverzeichnis und ENV
WORKDIR /app
ENV PORT=8080
EXPOSE 8080

# Starte die Anwendung
CMD ["run-app"]
