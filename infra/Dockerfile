# syntax=docker/dockerfile:1

# Build The Application From Source
FROM golang:1.26 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-strykz ./cmd/nana-boiler


##TODO: RUN TESTS

## Lean Image
FROM gcr.io/distroless/base-debian12 AS release-stage

WORKDIR /app

COPY --from=build-stage /docker-strykz /docker-strykz
COPY --from=build-stage /app/scripts /app/scripts

EXPOSE 8080
EXPOSE 6767



CMD ["/docker-strykz"]