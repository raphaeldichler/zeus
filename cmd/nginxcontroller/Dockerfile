FROM golang:1.24 AS builder

WORKDIR /opt/zeus/nginxcontroller
COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./nginxcontroller ./cmd/nginxcontroller

# nginx:1.27-alpine
FROM nginx@sha256:65645c7bb6a0661892a8b03b89d0743208a18dd2f3f17a54ef4b76fb8e2f2a10

COPY cmd/nginxcontroller/nginx.conf /etc/zeus/ingress/nginx.conf

COPY --from=builder /opt/zeus/nginxcontroller /opt/zeus/nginxcontroller
RUN chmod +x /opt/zeus/nginxcontroller

COPY cmd/nginxcontroller/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

RUN rm -rf /etc/nginx
ENTRYPOINT ["./entrypoint.sh"]
