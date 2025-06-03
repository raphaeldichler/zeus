FROM golang:1.24 AS builder

WORKDIR /opt/zeus/nginxcontroller
COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./nginxcontroller ./cmd/nginxcontroller

FROM nginx:1.27-alpine

COPY ./cmd/nginxcontroller/nginx.conf /etc/nginx/nginx.conf

COPY --from=builder /opt/zeus/nginxcontroller/nginxcontroller /opt/zeus/nginxcontroller
RUN chmod +x /opt/zeus/nginxcontroller

COPY cmd/nginxcontroller/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

RUN rm -rf /etc/nginx/conf.d/
ENTRYPOINT ["./entrypoint.sh"]
