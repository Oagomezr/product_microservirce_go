# Multi-stage: compile in a full Go image, run from a small runtime image.

# --- Stage 1: build ---
FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/products ./cmd/products

# --- Stage 2: runtime ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/products .

EXPOSE 8080

# Config comes from environment variables (see .env.example). Do not COPY .env into the image.
USER nobody:nobody

ENTRYPOINT ["./products"]
