# syntax=docker/dockerfile:1

# Build The Application From Source
FROM golang:1.26.0-bookworm AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-strykz ./cmd/nana-boiler

## Lean Image
FROM gcr.io/distroless/base-debian12 AS release-stage

ARG GIT_SHA=unknown
LABEL org.opencontainers.image.revision=$GIT_SHA

WORKDIR /app

COPY --from=build-stage /docker-strykz /docker-strykz

EXPOSE 8080
EXPOSE 6767

ENTRYPOINT ["/docker-strykz"]
