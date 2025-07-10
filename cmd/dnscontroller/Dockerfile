FROM golang:1.24 AS builder

WORKDIR /opt/zeus/dnscontroller
COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./dnscontroller ./cmd/dnscontroller

FROM coredns/coredns:latest

COPY Corefile /etc/coredns/Corefile

COPY --from=builder /opt/zeus/dnscontroller/dnscontroller /opt/zeus/dnscontroller
RUN chmod +x /opt/zeus/dnscontroller


