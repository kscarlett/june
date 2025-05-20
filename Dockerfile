# Use a minimal Go image for building
FROM golang:1.23-alpine AS build

WORKDIR /app
COPY . .
RUN go build -o june ./cmd/june

# Use a minimal runtime image
FROM alpine:latest

WORKDIR /site
COPY --from=build /app/june /usr/local/bin/june

# Entrypoint: expects input and output as arguments
ENTRYPOINT ["june", "generate"]