FROM golang:1.25-alpine AS build

ENV CGO_ENABLED=0
ARG version=undefined

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN go build -trimpath -o /out/viarticles -ldflags "-X main.version=${version} -s -w" ./cmd/viarticles

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    addgroup -S viarticles && adduser -S -G viarticles viarticles && \
    mkdir -p /data && chown viarticles:viarticles /data
COPY --from=build /out/viarticles /usr/local/bin/viarticles

USER viarticles
VOLUME ["/data"]
EXPOSE 8080
CMD ["viarticles"]
