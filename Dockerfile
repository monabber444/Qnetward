FROM golang:1.26-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/netwatcher ./cmd/netwatcher
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/probe-agent ./cmd/probe-agent
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/collector ./cmd/collector
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/host-agent ./cmd/host-agent

FROM debian:bookworm-slim

RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates iproute2 iputils-ping \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /out/netwatcher /app/netwatcher
COPY --from=build /out/probe-agent /app/probe-agent
COPY --from=build /out/collector /app/collector
COPY --from=build /out/host-agent /app/host-agent
COPY drivers/profiles /app/drivers/profiles

EXPOSE 8080 8081 9090 1162/udp 2055/udp 7946/udp 9000/udp

ENTRYPOINT ["/app/netwatcher", "serve"]
