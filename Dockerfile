FROM golang:1.14 AS build-env

RUN mkdir -p /build/onos-stress

WORKDIR /build/onos-stress
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
        -o /build/onos-stress/bin/onos-stress \
        main.go

# -------------
# Image creation stage

FROM alpine:latest

ENV ONOS_COMPONENT_LLDPLINKPROVIDER_CONFIG_FILE=/app/config/LldpLinkProvider.json
ENV ONOS_COMPONENT_OLTFLOWSERVICE_CONFIG_FILE=/app/config/OltFlowService.json
ENV ONOS_API_LOGIN=onos
ENV ONOS_API_PASS=rocks
ENV SLEEP_AFTER_FAILURE=1s
ENV SLEEP_AFTER_SUCCESS=10s

RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build-env /build/onos-stress/bin/onos-stress /app/
COPY config /app/config

CMD /app/onos-stress
